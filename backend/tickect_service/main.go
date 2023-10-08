package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lapayka/rsoi-lab2/ticket-service/DA"
)

type GateWay struct {
	db *DA.DB
	//logger *slog.Logger
}

func main() {
	router := mux.NewRouter()

	db, _ := DA.New("localhost", "postgres", "tickets", "1234")
	gw := GateWay{db}

	router.HandleFunc("/api/v1/tickets/{ticketUid}", gw.getTicketByUUID).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.getTicketsByUsername).Methods("Get")

	err := http.ListenAndServe(":8070", router)
	if err != nil {
		//gw.logger.Error("failed to run app", "error", err)
	}
}

func ReadSerializable(r *http.Request, item any) error {
	buff, _ := io.ReadAll(r.Body)

	err := json.Unmarshal(buff, item)

	return err
}

func WriteSerializable(item any, w http.ResponseWriter) {
	bytes, _ := json.Marshal(item)
	w.Write(bytes)
}

func (gw *GateWay) getTicketByUUIDAndUserName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketUid := vars["ticketUid"]
	username := r.Header.Get("X-User-Name")

	ticket, err := gw.db.GetTicketByUUID(ticketUid, username)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		WriteSerializable(ticket, w)
		w.WriteHeader(http.StatusOK)
	}
}

func (gw *GateWay) getTicketsByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	tickets, err := gw.db.GetTicketsByUsername(username)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		WriteSerializable(tickets, w)
		w.WriteHeader(http.StatusOK)
	}
}
