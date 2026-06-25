package port

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type TicketRepository interface {
	Create(
		tx *sql.Tx,
		reservationID int64,
		eventID int64,
		ticketNumber int,
		token string,
	) error

	FindByReservationID(
		reservationID int64,
	) ([]entity.Ticket, error)

	MarkCheckinIfValid(
		tx *sql.Tx,
		token string,
	) (int64, error)

	FindTicketViewByToken(
		token string,
	) (entity.TicketView, error)

	FindByToken(token string) (entity.Ticket, error)

	GetLastCheckinsByEventID(
		eventID int64,
		limit int,
	) ([]entity.LastCheckin, error)
}