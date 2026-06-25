package entity

import "time"

type Reservation struct {
	ID       int64
	EventID  int64
	Name     string
	Email    string
	Quantity int
	Status   string
	Token    string
	CheckedIn bool
	CreatedAt time.Time
}