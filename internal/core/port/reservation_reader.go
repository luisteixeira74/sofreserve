package port

type ReservationReader interface {
	SumConfirmedByEvent(eventID int) (int, error)
}