package port

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type ReservationRepository interface {
	SumByEventID(eventID int) (int, error)

	Create(
		tx *sql.Tx,
		eventID int,
		name string,
		email string,
		qty int,
		status string,
		token string,
	) error

	ExistsByEventAndEmail(
		tx *sql.Tx,
		eventID int,
		email string,
	) (bool, error)

	FindByTokenForUpdate(tx *sql.Tx, token string) (entity.Reservation, error)

	UpdateStatus(tx *sql.Tx, id int, status string) error
}