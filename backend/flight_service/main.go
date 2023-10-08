package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lapayka/rsoi-lab2/flight-service/DA"
)

type GateWay struct {
	db *DA.DB
	//logger *slog.Logger
}

func main() {
	router := mux.NewRouter()

	db, _ := DA.New("localhost", "postgres", "flights", "1234")
	gw := GateWay{db}

	router.HandleFunc("/api/v1/flights", gw.getFlights).Methods("Get")

	err := http.ListenAndServe(":8060", router)
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

func (gw *GateWay) getFlights(w http.ResponseWriter, r *http.Request) {
	flights, _ := gw.db.GetFlights()

	if len(flights) == 0 {
		w.WriteHeader(http.StatusNotFound)
	} else {
		WriteSerializable(flights, w)
		w.WriteHeader(http.StatusOK)
	}
}
