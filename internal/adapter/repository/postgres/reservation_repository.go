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

func (r *ReservationRepository) SumByEventID(eventID int) (int, error) {
	var total int

	err := r.db.QueryRow(
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

func (r *ReservationRepository) FindByTokenForUpdate(tx *sql.Tx, token string) (entity.Reservation, error) {
	var res entity.Reservation

	err := tx.QueryRow(`
		SELECT id, event_id, name, email, quantity, status, token
		FROM reservations
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(
		&res.ID,
		&res.EventID,
		&res.Name,
		&res.Email,
		&res.Quantity,
		&res.Status,
		&res.Token,
	)

	return res, err
}

func (r *ReservationRepository) UpdateStatus(
	tx *sql.Tx,
	id int,
	status string,
) error {

	_, err := tx.Exec(`
		UPDATE reservations
		SET status = $1
		WHERE id = $2
	`, status, id)

	return err
}