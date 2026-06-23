package usecase

import (
	"database/sql"
	"errors"
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

	tx, err := u.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	ticket, err := u.TicketRepo.FindByTokenForUpdate(
		tx,
		token,
	)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.ErrTicketNotFound
		}

		return err
	}

	if ticket.CheckedInAt != nil {
		return appErrors.ErrTicketAlreadyCheckedIn
	}

	err = u.TicketRepo.CheckIn(
		tx,
		ticket.ID,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}