package main

import (
	"net/http"

	"github.com/gorilla/mux"
	http_utils "github.com/lapayka/rsoi-2/Common/HTTP_Utils"
	PS_DA "github.com/lapayka/rsoi-2/privilege-service/da"
)

type GateWay struct {
	db *PS_DA.DB
	//logger *slog.Logger
}

func main() {
	router := mux.NewRouter()

	db, _ := PS_DA.New("localhost", "postgres", "bonus", "1234")
	gw := GateWay{db}

	router.HandleFunc("/api/v1/me", gw.getPrivilegeAndHistory).Methods("Get")
	router.HandleFunc("/api/v1/tickets", gw.getHistory).Methods("Get")

	err := http.ListenAndServe(":8070", router)
	if err != nil {
		//gw.logger.Error("failed to run app", "error", err)
	}
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
