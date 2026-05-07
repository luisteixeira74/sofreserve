package entity

import "time"

type Event struct {
	ID         int
	Name       string
	TotalSeats int
	EndsAt     time.Time
	PublicID   string
}

func CanTransition(from, to ReservationStatus) bool {
	switch from {
	case StatusPending:
		return to == StatusConfirmed || to == StatusCanceled
	case StatusConfirmed:
		return to == StatusCanceled
	case StatusCanceled:
		return false
	default:
		return false
	}
}