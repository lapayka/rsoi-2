package DA

import (
	"fmt"
	"log"

	"github.com/lapayka/rsoi-lab2/ticket-service/structs"
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

func (db *DB) GetTicketByUUID(uuid, username string) (structs.Ticket, error) {
	ticket := structs.Ticket{TicketUuid: uuid, Username: username}

	err := db.db.First(&ticket).Error

	return ticket, err
}

func (db *DB) GetTicketsByUsername(username string) (structs.Tickets, error) {
	tickets := structs.Tickets{}

	err := db.db.Find(&tickets).Where(&structs.Ticket{Username: username}).Error

	return tickets, err
}
