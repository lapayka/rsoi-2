package DTO

type BuyTicketDTO struct {
	FlightNumber    string `json: flightNumber`
	Price           int64  `json: price`
	PaidFromBalance bool   `json: paidFromBalance`
	TicketUid       string `json: TicketUuid,omitempty`
}
