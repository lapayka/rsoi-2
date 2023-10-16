package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	DTO "github.com/lapayka/rsoi-2/Common"
	http_utils "github.com/lapayka/rsoi-2/Common/HTTP_Utils"
	"github.com/lapayka/rsoi-2/Common/Logger"
	PS_DA "github.com/lapayka/rsoi-2/privilege-service/da"
	PS_structs "github.com/lapayka/rsoi-2/privilege-service/structs"
)

type GateWay struct {
	db *PS_DA.DB
	//logger *slog.Logger
}

func main() {
	router := mux.NewRouter()

	db, _ := PS_DA.New("localhost", "postgres", "bonus", "1234")
	gw := GateWay{db}

	router.HandleFunc("/manage/health", http_utils.HealthCkeck).Methods("Get")
	router.HandleFunc("/api/v1/me", gw.getPrivilegeAndHistory).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.getHistory).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.buyTicket).Methods("Post")

	err := http.ListenAndServe(":8050", router)
	if err != nil {
		//gw.logger.Error("failed to run app", "error", err)
	}
}

func (gw *GateWay) buyTicket(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	buy_ticket_info := DTO.BuyTicketDTO{}
	http_utils.ReadSerializable(r, &buy_ticket_info)

	privelege_item := PS_structs.Privilege_history{DateTime: time.Now(), TicketUID: buy_ticket_info.TicketUid}

	privelege_item.OperationType = "DEBIT_THE_ACCOUNT"
	if buy_ticket_info.PaidFromBalance {
		privelege_item.OperationType = "FILL_IN_BALANCE"
	}
	err := gw.db.CreateTicket(username, buy_ticket_info.Price, buy_ticket_info.PaidFromBalance, privelege_item)

	if err != nil {
		Logger.GetLogger().Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (gw *GateWay) getPrivilegeAndHistory(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	p, err := gw.db.GetPrivilegeAndHistoryByUserName(username)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http_utils.WriteSerializable(p, w)
}

func (gw *GateWay) getHistory(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User-Name")

	p, err := gw.db.GetPrivilegeAndHistoryByUserName(username)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http_utils.WriteSerializable(p.History, w)
}
