package port

type ReservationRepository interface {
	ExistsByEventAndEmail(eventID int, email string) (bool, error)

	Create(eventID int, name, email string, quantity int, status, token string) error

	SumConfirmedByEvent(eventID int) (int, error)
}