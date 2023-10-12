package TS_structs

type Ticket struct {
	ID           int64  `json:"id"`
	TicketUuid   string `json:"ticket_uuid"`
	Username     string `json:"username"`
	FlightNumber string `json:"flight_number"`
	Price        int64  `json:"price"`
	Status       string `json:"status"`
}

func (Ticket) TableName() string {
	return "ticket"
}

type Tickets []Ticket

func (Tickets) TableName() string {
	return "ticket"
}
