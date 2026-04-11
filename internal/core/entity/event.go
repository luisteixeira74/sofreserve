package entity

import "time"

type Event struct {
	ID         int
	Name       string
	TotalSeats int
	EndsAt     time.Time
}