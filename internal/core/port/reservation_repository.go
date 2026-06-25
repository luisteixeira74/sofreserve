package port

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type ReservationRepository interface {
	SumByEventID(eventID int64) (int, error)

	Create(
		tx *sql.Tx,
		eventID int64,
		name string,
		email string,
		qty int,
		status string,
		token string,
	) (int64, error)

	ExistsByEventAndEmail(
		tx *sql.Tx,
		eventID int64,
		email string,
	) (bool, error)

	FindByTokenForUpdate(tx *sql.Tx, token string) (entity.Reservation, error)

	UpdateStatus(tx *sql.Tx, id int64, status string) error
	
	FindConfirmedByEventID(eventID int64) ([]entity.Reservation, error)
}