package FS_structs

import (
	"time"
)

type Flight struct {
	ID           int64     `json:"id"`
	FlightNumber string    `json:"flight_number"`
	Date         time.Time `json:"date"`
	FromAirport  Airport   `json:"from_airport"`
	ToAirport    Airport   `json:"to_airport"`
}

type Flights = []Flight
