package postgres

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) SumByEventID(tx *sql.Tx, eventID int) (int, error) {
	var total int

	err := tx.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) 
		 FROM reservations 
		 WHERE event_id = $1 AND status = $2`,
		eventID,
		string(entity.StatusConfirmed),
	).Scan(&total)

	return total, err
}

func (r *ReservationRepository) Create(
	tx *sql.Tx,
	eventID int,
	name string,
	email string,
	qty int,
	status string,
	token string,
) error {

	_, err := tx.Exec(
		`INSERT INTO reservations 
		(event_id, name, email, quantity, status, token) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		eventID, name, email, qty, status, token,
	)

	return err
}

func (r *ReservationRepository) ExistsByEventAndEmail(
	tx *sql.Tx,
	eventID int,
	email string,
) (bool, error) {

	var exists bool

	err := tx.QueryRow(
		`SELECT EXISTS(
			SELECT 1 
			FROM reservations 
			WHERE event_id = $1 
			AND email = $2 
			AND status IN ($3, $4)
		)`,
		eventID,
		email,
		string(entity.StatusPending),
		string(entity.StatusConfirmed),
	).Scan(&exists)

	return exists, err
}