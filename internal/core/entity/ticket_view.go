package entity

import "time"

type TicketView struct {
	EventName    string
	Token        string
	TicketNumber int64
	CheckedInAt  *time.Time
}