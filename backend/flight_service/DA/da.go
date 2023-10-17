package FS_DA

import (
	"fmt"
	"log"
	"time"

	FS_structs "github.com/lapayka/rsoi-2/flight_service/Structs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func New(host, user, db_name, password string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s", host, user, db_name, password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("unable to connect database", err)
	}

	return &DB{db: db}, nil
}

type joinRes struct {
	ID           int64
	FlightNumber string
	Date         time.Time

	FromAirportID      int64
	FromAirportName    string
	FromAirportCity    string
	FromAirportCountry string

	ToAirportID      int64
	ToAirportName    string
	ToAirportCity    string
	ToAirportCountry string
}

func (d *DB) GetFlights() (FS_structs.Flights, error) {
	flights := FS_structs.Flights{}

	joinres := []joinRes{}
	d.db.Table("flight").Select("flight.id, flight.flight_number, flight.datetime as date, fa.id as From_Airport_ID, fa.name as From_Airport_Name, fa.city as From_Airport_City, fa.country as From_Airport_Country, ta.id as To_Airport_ID, ta.name as To_Airport_Name, ta.city as To_Airport_City, ta.country as To_Airport_Country").
		Joins("JOIN airport fa on flight.from_airport_id = fa.id").
		Joins("JOIN airport ta on flight.to_airport_id = ta.id").
		Scan(&joinres)

	for _, res := range joinres {
		flights = append(flights, FS_structs.Flight{ID: res.ID, FlightNumber: res.FlightNumber, Date: res.Date, FromAirport: FS_structs.Airport{ID: res.FromAirportID, Name: res.FromAirportName, City: res.FromAirportCity, Country: res.FromAirportCountry}, ToAirport: FS_structs.Airport{ID: res.ToAirportID, Name: res.ToAirportName, City: res.ToAirportCity, Country: res.ToAirportCountry}})
	}

	if len(flights) == 0 {
		return nil, nil
	}

	return flights, nil
}
