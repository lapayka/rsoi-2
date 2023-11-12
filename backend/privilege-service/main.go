package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	DTO "github.com/lapayka/rsoi-2/Common"
	http_utils "github.com/lapayka/rsoi-2/Common/HTTP_Utils"
	"github.com/lapayka/rsoi-2/Common/Logger"
	PS_DA "github.com/lapayka/rsoi-2/privilege-service/da"
	PS_structs "github.com/lapayka/rsoi-2/privilege-service/structs"
	TS_structs "github.com/lapayka/rsoi-2/tickect_service/structs"

	amqp "github.com/rabbitmq/amqp091-go"
)

type GateWay struct {
	db *PS_DA.DB
	//logger *slog.Logger

	rabbit_conn  *amqp.Connection
	rabbit_chan  *amqp.Channel
	rabbit_queue amqp.Queue
}

func main() {
	router := mux.NewRouter()

	db, _ := PS_DA.New("localhost", "postgres", "bonus", "1234")

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

	gw := GateWay{db: db,
		rabbit_conn:  conn,
		rabbit_chan:  ch,
		rabbit_queue: q}

	router.HandleFunc("/manage/health", http_utils.HealthCkeck).Methods("Get")
	router.HandleFunc("/api/v1/me", gw.getPrivilegeAndHistory).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.getHistory).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.buyTicket).Methods("Post")
	//router.HandleFunc("/api/v1/tickets/{ticketUid}", gw.deleteTicket).Methods("DELETE")

	gw.deleteTicket()

	err = http.ListenAndServe(":8050", router)
	if err != nil {
		//gw.logger.Error("failed to run app", "error", err)
	}
}

func (gw *GateWay) deleteTicket() {
	fmt.Println("-->deleteTicket")

	msgs, err := gw.rabbit_chan.Consume(gw.rabbit_queue.Name, "", true, false, false, false, nil)
	if err != nil {
		Logger.GetLogger().Print(err)
	}

	fmt.Println("Consume")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			ticket := TS_structs.Ticket{}
			_ = json.Unmarshal(d.Body, &ticket)

			fmt.Println(string(d.Body))

			err := gw.db.DeleteTicket(ticket.TicketUid, ticket.Price)
			if err != nil {
				Logger.GetLogger().Println(err)
				return
			}
		}
	}()

	<-forever
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
