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

func (r *ReservationRepository) SumByEventID(eventID int64) (int, error) {
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
	eventID int64,
	name string,
	email string,
	qty int,
	status string,
	token string,
) (int64, error) {

	var id int64

	err := tx.QueryRow(`
		INSERT INTO reservations
		(event_id, name, email, quantity, status, token)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		eventID, name, email, qty, status, token,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *ReservationRepository) ExistsByEventAndEmail(
	tx *sql.Tx,
	eventID int64,
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
	id int64,
	status string,
) error {

	query := `
		UPDATE 	reservations
		SET
			status = $1,
			confirmed_at = CASE
				WHEN $1 = 'confirmed' THEN NOW()
				ELSE confirmed_at
			END,
			canceled_at = CASE
				WHEN $1 = 'canceled' THEN NOW()
				ELSE canceled_at
			END
		WHERE id = $2
	`

	_, err := tx.Exec(query, status, id)

	return err
}

func (r *ReservationRepository) FindConfirmedByEventID(eventID int64) ([]entity.Reservation, error) {
	rows, err := r.db.Query(`
		SELECT
			id,
			event_id,
			name,
			email,
			quantity,
			token,
			status,
			created_at
		FROM reservations
		WHERE event_id = $1
		AND status = 'confirmed'
		ORDER BY created_at ASC
	`, eventID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []entity.Reservation

	for rows.Next() {
		var reservation entity.Reservation

		err := rows.Scan(
			&reservation.ID,
			&reservation.EventID,
			&reservation.Name,
			&reservation.Email,
			&reservation.Quantity,
			&reservation.Token,
			&reservation.Status,
			&reservation.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}