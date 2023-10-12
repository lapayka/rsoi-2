package main

import (
	"net/http"

	"github.com/gorilla/mux"
	http_utils "github.com/lapayka/rsoi-2/Common/HTTP_Utils"
	FS_DA "github.com/lapayka/rsoi-2/flight_service/DA"
)

type GateWay struct {
	db *FS_DA.DB
	//logger *slog.Logger
}

func main() {
	router := mux.NewRouter()

	db, _ := FS_DA.New("localhost", "postgres", "flights", "1234")
	gw := GateWay{db}

	router.HandleFunc("/api/v1/flights", gw.getFlights).Methods("Get")

	err := http.ListenAndServe(":8060", router)
	if err != nil {
		//gw.logger.Error("failed to run app", "error", err)
	}
}

func (gw *GateWay) getFlights(w http.ResponseWriter, r *http.Request) {
	flights, _ := gw.db.GetFlights()

	if len(flights) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		http_utils.WriteSerializable(flights, w)
		w.WriteHeader(http.StatusOK)
	}
}
