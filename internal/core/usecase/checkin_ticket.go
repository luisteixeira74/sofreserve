package usecase

import (
	"database/sql"
	appErrors "sof-reserve/internal/core/errors"
	"sof-reserve/internal/core/port"
)

type CheckinTicket struct {
	DB         *sql.DB
	TicketRepo port.TicketRepository
}

func NewCheckinTicket(
    db *sql.DB,
    ticketRepo port.TicketRepository,
) *CheckinTicket {

    return &CheckinTicket{
        DB: db,
        TicketRepo: ticketRepo,
    }
}

func (u *CheckinTicket) Execute(token string) error {
	if token == "" {
		return appErrors.ErrInvalidToken
	}

	tx, err := u.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = u.TicketRepo.CheckIn(tx, token)
	if err != nil {
		return err
	}

	return tx.Commit()
}