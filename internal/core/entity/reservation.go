package entity

import "time"

type Reservation struct {
	ID       int
	EventID  int
	Name     string
	Email    string
	Quantity int
	Status   string
	Token    string
	CheckedIn bool
	CreatedAt time.Time
}