package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	DTO "github.com/lapayka/rsoi-2/Common"
	http_utils "github.com/lapayka/rsoi-2/Common/HTTP_Utils"
	"github.com/lapayka/rsoi-2/Common/Logger"
	TS_DA "github.com/lapayka/rsoi-2/tickect_service/DA"
	TS_structs "github.com/lapayka/rsoi-2/tickect_service/structs"
)

type GateWay struct {
	db *TS_DA.DB
	//logger *slog.Logger
}

func main() {
	router := mux.NewRouter()

	db, err := TS_DA.New("localhost", "postgres", "tickets", "1234")
	if err != nil {
		Logger.GetLogger().Print(err)
		return
	}

	gw := GateWay{db}

	router.HandleFunc("/manage/health", http_utils.HealthCkeck).Methods("Get")
	router.HandleFunc("/api/v1/tickets/{ticketUid}", gw.getTicketByUUIDAndUserName).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.getTicketsByUsername).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.createTicket).Methods("Post")
	router.HandleFunc("/api/v1/tickets/{ticketUid}", gw.deleteTicket).Methods("DELETE")

	err = http.ListenAndServe(":8070", router)
	if err != nil {
		Logger.GetLogger().Print(err)
	}
}

func (gw *GateWay) deleteTicket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketUid := vars["ticketUid"]
	username := r.Header.Get("X-User-Name")
	if username == "" || ticketUid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ticket := TS_structs.Ticket{Username: username, TicketUid: ticketUid}

	err := gw.db.DeleteTicket(&ticket)

	if err != nil {
		Logger.GetLogger().Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http_utils.WriteSerializable(&ticket, w)
	w.WriteHeader(http.StatusOK)
}

func (gw *GateWay) createTicket(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	buy_ticket_info := DTO.BuyTicketDTO{}
	http_utils.ReadSerializable(r, &buy_ticket_info)

	ticket := TS_structs.Ticket{TicketUid: uuid.New().String(), Username: username, FlightNumber: buy_ticket_info.FlightNumber, Price: buy_ticket_info.Price, Status: "PAID"}

	err := gw.db.CreateTicket(&ticket)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http_utils.WriteSerializable(ticket, w)
	w.WriteHeader(http.StatusCreated)
}

func (gw *GateWay) getTicketByUUIDAndUserName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketUid := vars["ticketUid"]
	username := r.Header.Get("X-User-Name")

	ticket, err := gw.db.GetTicketByUUID(ticketUid, username)

	if err != nil {
		Logger.GetLogger().Print(err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		http_utils.WriteSerializable(ticket, w)
		w.WriteHeader(http.StatusOK)
	}
}

func (gw *GateWay) getTicketsByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	tickets, err := gw.db.GetTicketsByUsername(username)

	if err != nil {
		Logger.GetLogger().Print(err)
		w.WriteHeader(http.StatusNotFound)
	} else {
		http_utils.WriteSerializable(tickets, w)
		w.WriteHeader(http.StatusOK)
	}
}
