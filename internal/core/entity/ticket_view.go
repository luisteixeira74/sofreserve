package entity

import "time"

type TicketView struct {
	EventName    string
	Token        string
	TicketNumber int
	CheckedInAt  *time.Time
}