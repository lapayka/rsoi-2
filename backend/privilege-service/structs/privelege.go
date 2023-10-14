package PS_structs

import (
	"time"
)

type Privilege struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
	Balance  int64  `json:"balance"`
}

type Privilege_history struct {
	ID            int64     `json:"id"`
	PrivilegeID   int64     `json:"Privilege_id"`
	TicketUID     string    `json:"ticket_uid"`
	DateTime      time.Time `json:"datetime" gorm:"column:datetime"`
	BalanceDiff   int64     `json:"balance_diff"`
	OperationType string    `json:"operation_type"`
}

type Privileges_history = []Privilege_history

type Privilege_with_history struct {
	Privilege_info Privilege
	History        Privileges_history
}

type Tabler interface {
	TableName() string
}

func (Privilege) TableName() string {
	return "privilege"
}

func (Privilege_history) TableName() string {
	return "privilege_history"
}
