package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/flights", flight_proxy).Methods("Get")
	router.HandleFunc("/api/v1/me", bonus_proxy).Methods("Get")
	router.HandleFunc("/api/v1/tickets", bonus_proxy).Methods("Get")
	router.HandleFunc("/api/v1/tickets/{ticketUid}", ticket_proxy).Methods("Get")
	router.HandleFunc("/api/v1/tickets", ticket_proxy).Methods("Get")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		//gw.logger.Error("failed to run app", "error", err)
	}
}

func echo_request(w http.ResponseWriter, r *http.Request, service_url string) {
	req, _ := http.NewRequest(r.Method, service_url+r.URL.String(), r.Body)
	req.Header = r.Header
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
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
