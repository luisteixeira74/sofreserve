package port

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type TicketRepository interface {
	Create(
		tx *sql.Tx,
		reservationID int64,
		ticketNumber int,
		token string,
	) error

	FindByReservationID(
		reservationID int64,
	) ([]entity.Ticket, error)

	FindByToken(
		token string,
	) (entity.Ticket, error)

	FindByTokenForUpdate(
		tx *sql.Tx,
		token string,
	) (entity.Ticket, error)

	CheckIn(
		tx *sql.Tx,
		ticketID int64,
	) error
}