package postgres

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) GetByID(id int) (entity.Event, error) {
	var e entity.Event

	err := r.db.QueryRow(
		`SELECT id, name, total_seats, ends_at, public_id
		 FROM events
		 WHERE id = $1`,
		id,
	).Scan(
		&e.ID,
		&e.Name,
		&e.TotalSeats,
		&e.EndsAt,
		&e.PublicID,
	)

	return e, err
}

func (r *EventRepository) GetByPublicID(publicID string) (entity.Event, error) {
	var e entity.Event

	err := r.db.QueryRow(
		`SELECT id, name, total_seats, ends_at, public_id
		 FROM events
		 WHERE public_id = $1`,
		publicID,
	).Scan(
		&e.ID,
		&e.Name,
		&e.TotalSeats,
		&e.EndsAt,
		&e.PublicID,
	)

	return e, err
}

func (r *EventRepository) FindByIDForUpdate(tx *sql.Tx, id int) (entity.Event, error) {
	var e entity.Event

	err := tx.QueryRow(
		`SELECT id, name, total_seats, ends_at, public_id
		 FROM events
		 WHERE id = $1
		 FOR UPDATE`,
		id,
	).Scan(
		&e.ID,
		&e.Name,
		&e.TotalSeats,
		&e.EndsAt,
		&e.PublicID,
	)

	return e, err
}