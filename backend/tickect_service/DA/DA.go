package TS_DA

import (
	"fmt"
	"log"

	TS_structs "github.com/lapayka/rsoi-2/tickect_service/structs"
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

func (db *DB) GetTicketByUUID(uuid, username string) (TS_structs.Ticket, error) {
	ticket := TS_structs.Ticket{TicketUuid: uuid, Username: username}

	err := db.db.First(&ticket).Error

	return ticket, err
}

func (db *DB) GetTicketsByUsername(username string) (TS_structs.Tickets, error) {
	tickets := TS_structs.Tickets{}

	err := db.db.Find(&tickets).Where(&TS_structs.Ticket{Username: username}).Error

	return tickets, err
}
