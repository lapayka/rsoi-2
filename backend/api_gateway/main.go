package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	DTO "github.com/lapayka/rsoi-2/Common"
	http_utils "github.com/lapayka/rsoi-2/Common/HTTP_Utils"
	"github.com/lapayka/rsoi-2/Common/Logger"
	FS_structs "github.com/lapayka/rsoi-2/flight_service/Structs"
	TS_structs "github.com/lapayka/rsoi-2/tickect_service/structs"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/manage/health", http_utils.HealthCkeck).Methods("Get")
	router.HandleFunc("/api/v1/flights", flight_proxy).Methods("Get")
	router.HandleFunc("/api/v1/me", bonus_proxy).Methods("Get")
	router.HandleFunc("/api/v1/tickets/{ticketUid}", ticket_proxy).Methods("Get")
	router.HandleFunc("/api/v1/tickets", ticket_proxy).Methods("Get")
	router.HandleFunc("/api/v1/tickets/{ticketUid}", ticket_proxy).Methods("DELETE")

	router.HandleFunc("/api/v1/tickets", buy_ticket).Methods("Post")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		Logger.GetLogger().Print(err)
	}
}

func GetDefaultClient() *http.Client {
	client := http.DefaultClient
	client.Timeout = 2 * time.Second

	return client
}

func check_flght_number(flight_number string) bool {
	req, _ := http.NewRequest("GET", "http://localhost:8060/api/v1/flights", nil)
	r, err := GetDefaultClient().Do(req)

	if err != nil {
		Logger.GetLogger().Print(err)
		return false
	}

	flights := FS_structs.Flights{}
	err = http_utils.ReadSerializableFromResponse(r, &flights)

	if err != nil {
		Logger.GetLogger().Print(err)
		return false
	}

	for i := range flights {
		if flights[i].FlightNumber == flight_number {
			return true
		}
	}

	return false
}

func buy_ticket_in_ticket_service(username string, buy_ticket_info DTO.BuyTicketDTO) (TS_structs.Ticket, error) {
	body, _ := json.Marshal(buy_ticket_info)
	reader := strings.NewReader(string(body))

	req, err := http.NewRequest("POST", "http://localhost:8070/api/v1/tickets", reader)
	req.Header.Add("X-User-Name", username)

	if err != nil {
		Logger.GetLogger().Print(err)
		return TS_structs.Ticket{}, err
	}

	var r *http.Response
	r, err = GetDefaultClient().Do(req)

	if err != nil {
		Logger.GetLogger().Print(err)
		return TS_structs.Ticket{}, err
	}

	ticket := TS_structs.Ticket{}
	err = http_utils.ReadSerializableFromResponse(r, &ticket)

	if err != nil {
		Logger.GetLogger().Print(err)
		return TS_structs.Ticket{}, err
	}

	return ticket, nil
}

func buy_ticket_in_privilege_service(username string, buy_ticket_info DTO.BuyTicketDTO) error {
	body, _ := json.Marshal(buy_ticket_info)
	reader := strings.NewReader(string(body))

	req, err := http.NewRequest("POST", "http://localhost:8050/api/v1/tickets", reader)
	req.Header.Add("X-User-Name", username)

	if err != nil {
		Logger.GetLogger().Print(err)
		return err
	}

	var r *http.Response
	r, err = GetDefaultClient().Do(req)

	if err != nil {
		Logger.GetLogger().Print(err)
		return err
	}

	if r.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("status code was: %d\n", r.StatusCode)
}

func buy_ticket(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	buy_ticket_info := DTO.BuyTicketDTO{}
	http_utils.ReadSerializable(r, &buy_ticket_info)

	if !check_flght_number(buy_ticket_info.FlightNumber) {
		w.WriteHeader(http.StatusNotFound)
	}

	ticket, err := buy_ticket_in_ticket_service(username, buy_ticket_info)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	buy_ticket_info.TicketUid = ticket.TicketUid

	err = buy_ticket_in_privilege_service(username, buy_ticket_info)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func echo_request(w http.ResponseWriter, r *http.Request, service_url string) {
	req, _ := http.NewRequest(r.Method, service_url+r.URL.String(), r.Body)
	fmt.Println(r.Method)
	req.Header = r.Header
	response, err := GetDefaultClient().Do(req)

	if err != nil {
		Logger.GetLogger().Print(err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(response.StatusCode)
		io.Copy(w, response.Body)
	}
}

func bonus_proxy(w http.ResponseWriter, r *http.Request) {
	echo_request(w, r, "http://localhost:8050")
}

func flight_proxy(w http.ResponseWriter, r *http.Request) {
	echo_request(w, r, "http://localhost:8060")
}

func ticket_proxy(w http.ResponseWriter, r *http.Request) {
	echo_request(w, r, "http://localhost:8070")
}
