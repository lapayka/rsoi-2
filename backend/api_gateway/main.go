package main

import (
	"container/list"
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
	"github.com/lapayka/rsoi-2/api_gateway/circuit_breaker"
	FS_structs "github.com/lapayka/rsoi-2/flight_service/Structs"
	TS_structs "github.com/lapayka/rsoi-2/tickect_service/structs"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Query struct {
	Service *circuit_breaker.CircuitBreaker
	Request http.Request
}

type GateWay struct {
	ticket_service    circuit_breaker.CircuitBreaker
	privilege_service circuit_breaker.CircuitBreaker
	flight_service    circuit_breaker.CircuitBreaker

	Queue *list.List

	rabbit_conn  *amqp.Connection
	rabbit_chan  *amqp.Channel
	rabbit_queue amqp.Queue
}

func (gw *GateWay) CleanQuery() {
	queue_tmp := list.New()
	Logger.GetLogger().Println("Enter into delayed requests")

	for gw.Queue.Len() > 0 {
		e := gw.Queue.Front()
		query := Query(e.Value.(Query))

		fmt.Printf("Trying to %s %s%s\n", query.Request.Method, query.Service.Host, query.Request.URL.String())
		resp, err := query.Service.SendQuery(&query.Request)

		if err != nil || resp.StatusCode != http.StatusInternalServerError {
			queue_tmp.PushBack(query)
		}
		gw.Queue.Remove(e)
		fmt.Println(gw.Queue.Len())
	}

	gw.Queue = queue_tmp
}

func SetTimer(gw *GateWay) {
	timer := time.NewTimer(10 * time.Second)
	go func() {
		<-timer.C
		gw.CleanQuery()
		SetTimer(gw)
	}()
}

func main() {
	router := mux.NewRouter()

	_privelege_service := circuit_breaker.CircuitBreaker{Host: "http://localhost:8050"}
	_flight_service := circuit_breaker.CircuitBreaker{Host: "http://localhost:8060"}
	_ticket_service := circuit_breaker.CircuitBreaker{Host: "http://localhost:8070"}

	// Rabbit init
	// -------------------------------------------------------------------
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		Logger.GetLogger().Print(err)
		defer conn.Close()
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		Logger.GetLogger().Print(err)
		defer ch.Close()
		return
	}

	q, err := ch.QueueDeclare("tickets", false, false, false, false, nil)
	if err != nil {
		Logger.GetLogger().Print(err)
		return
	}
	// -------------------------------------------------------------------

	gw := GateWay{ticket_service: _ticket_service,
		privilege_service: _privelege_service,
		flight_service:    _flight_service,
		Queue:             list.New(),
		rabbit_conn:       conn,
		rabbit_chan:       ch,
		rabbit_queue:      q}

	router.HandleFunc("/manage/health", http_utils.HealthCkeck).Methods("Get")
	router.HandleFunc("/api/v1/flights", gw.flight_service.ProxyQuery).Methods("Get")
	router.HandleFunc("/api/v1/me", gw.privilege_service.ProxyQuery).Methods("Get")
	router.HandleFunc("/api/v1/tickets/{ticketUid}", gw.ticket_service.ProxyQuery).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.ticket_service.ProxyQuery).Methods("Get")

	router.HandleFunc("/api/v1/tickets/{ticketUid}", gw.delete_ticket).Methods("DELETE")
	router.HandleFunc("/api/v1/tickets", gw.buy_ticket).Methods("Post")

	//SetTimer(&gw)

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		Logger.GetLogger().Print(err)
	}
}

func GetDefaultClient() *http.Client {
	client := http.DefaultClient
	client.Timeout = 2 * time.Second

	return client
}

func (gw *GateWay) delete_ticket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-->delete_ticket")
	fmt.Println(r.URL.String())
	req, _ := http.NewRequest("DELETE", r.URL.String(), nil)
	req.Header.Set("X-User-Name", r.Header.Get("X-User-Name"))
	resp, err := gw.ticket_service.SendQuery(req)
	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ticket := TS_structs.Ticket{}
	http_utils.ReadSerializableFromResponse(resp, &ticket)
	tmp, _ := json.Marshal(ticket)
	fmt.Println(string(tmp))
	gw.rabbit_chan.Publish("", gw.rabbit_queue.Name, false, false, amqp.Publishing{
		Body: tmp,
	})
}

func (gw *GateWay) check_flght_number(flight_number string) bool {
	req, _ := http.NewRequest("GET", "/api/v1/flights", nil)
	var r *http.Response
	r, err := gw.flight_service.SendQuery(req)

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

func (gw *GateWay) buy_ticket_in_ticket_service(username string, buy_ticket_info DTO.BuyTicketDTO) (TS_structs.Ticket, error) {
	body, _ := json.Marshal(buy_ticket_info)
	reader := strings.NewReader(string(body))

	req, err := http.NewRequest("POST", "/api/v1/tickets", reader)
	req.Header.Add("X-User-Name", username)

	if err != nil {
		Logger.GetLogger().Print(err)
		return TS_structs.Ticket{}, err
	}

	var r *http.Response
	r, err = gw.ticket_service.SendQuery(req)

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

func (gw *GateWay) buy_ticket_in_privilege_service(username string, buy_ticket_info DTO.BuyTicketDTO) error {
	body, _ := json.Marshal(buy_ticket_info)
	reader := strings.NewReader(string(body))

	req, err := http.NewRequest("POST", "/api/v1/tickets", reader)
	req.Header.Add("X-User-Name", username)

	if err != nil {
		Logger.GetLogger().Print(err)
		return err
	}

	var r *http.Response
	r, err = gw.privilege_service.SendQuery(req)

	if err != nil {
		Logger.GetLogger().Print(err)
		return err
	}

	if r.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("status code was: %d\n", r.StatusCode)
}

func (gw *GateWay) buy_ticket(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	buy_ticket_info := DTO.BuyTicketDTO{}
	http_utils.ReadSerializable(r, &buy_ticket_info)

	if !gw.check_flght_number(buy_ticket_info.FlightNumber) {
		w.WriteHeader(http.StatusNotFound)
	}

	ticket, err := gw.buy_ticket_in_ticket_service(username, buy_ticket_info)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	buy_ticket_info.TicketUid = ticket.TicketUid

	err = gw.buy_ticket_in_privilege_service(username, buy_ticket_info)

	if err != nil {
		req, _ := http.NewRequest("DELETE", "/api/v1/tickets/"+ticket.TicketUid, nil)
		req.Header.Add("X-User-Name", username)
		gw.ticket_service.SendQuery(req)

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
