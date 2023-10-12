package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	http_utils "github.com/lapayka/rsoi-2/Common/HTTP_Utils"
	TS_DA "github.com/lapayka/rsoi-2/tickect_service/DA"
)

type GateWay struct {
	db *TS_DA.DB
	//logger *slog.Logger
}

func main() {
	router := mux.NewRouter()

	db, _ := TS_DA.New("localhost", "postgres", "tickets", "1234")
	gw := GateWay{db}

	router.HandleFunc("/api/v1/tickets/{ticketUid}", gw.getTicketByUUIDAndUserName).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.getTicketsByUsername).Methods("Get")

	err := http.ListenAndServe(":8070", router)
	if err != nil {
		//gw.logger.Error("failed to run app", "error", err)
	}
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
		http_utils.WriteSerializable(ticket, w)
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
		http_utils.WriteSerializable(tickets, w)
		w.WriteHeader(http.StatusOK)
	}
}
