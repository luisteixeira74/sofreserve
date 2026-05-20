package entity

import "time"

type Ticket struct {
	ID             int64
	ReservationID  int64
	TicketNumber   int
	Token          string
	CheckedInAt    *time.Time
	CreatedAt      time.Time
}