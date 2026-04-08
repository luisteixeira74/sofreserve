package postgres

import "database/sql"

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) FindByIDForUpdate(tx *sql.Tx, id int) (int, int, error) {
	var totalSeats int

	err := tx.QueryRow(
		"SELECT total_seats FROM events WHERE id = $1 FOR UPDATE",
		id,
	).Scan(&totalSeats)

	return id, totalSeats, err
}

func (r *EventRepository) DB() *sql.DB {
	return r.db
}