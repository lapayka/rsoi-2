package DA

import (
	"fmt"
	"log"
	"time"

	"github.com/lapayka/rsoi-2/flight-service/Structs"
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

func (d *DB) GetFlights() (Structs.Flights, error) {
	flights := Structs.Flights{}

	joinres := []joinRes{}
	d.db.Table("flight").Select("flight.id, flight.flight_number, flight.datetime, fa.id, fa.name, fa.city, fa.country, ta.id, ta.name, ta.city, ta.country").Joins("JOIN airport fa on flight.from_airport_id = fa.id").Joins("JOIN airport ta on flight.to_airport_id = ta.id").Scan(&joinres)

	for _, res := range joinres {
		flights = append(flights, Structs.Flight{res.ID, res.FlightNumber, res.Date, Structs.Airport{res.FromAirportID, res.FromAirportName, res.FromAirportCity, res.FromAirportCountry}, Structs.Airport{res.ToAirportID, res.ToAirportName, res.ToAirportCity, res.ToAirportCountry}})
	}

	if len(flights) == 0 {
		return nil, nil
	}

	return flights, nil
}
