package DTO

type BuyTicketDTO struct {
	FlightNumber    string `json: FlightNumber`
	Price           int64  `json: price`
	PaidFromBalance bool   `json: PaidFromBalance`
	TicketUid       string `json:"ticket_uuid"`
}
