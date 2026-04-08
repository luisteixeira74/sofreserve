package postgres

import "database/sql"

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) SumByEventID(tx *sql.Tx, eventID int) (int, error) {
	var total int

	err := tx.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM reservations WHERE event_id = $1",
		eventID,
	).Scan(&total)

	return total, err
}

func (r *ReservationRepository) Create(tx *sql.Tx, eventID int, name string, qty int) error {
	_, err := tx.Exec(
		"INSERT INTO reservations (event_id, name, quantity) VALUES ($1, $2, $3)",
		eventID, name, qty,
	)

	return err
}