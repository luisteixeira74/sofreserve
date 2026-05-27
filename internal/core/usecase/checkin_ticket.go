package usecase

import (
	"database/sql"
	"errors"
	"sof-reserve/internal/core/port"
)

var ErrTicketAlreadyCheckedIn = errors.New("ticket already checked-in")

type CheckinTicket struct {
	DB         *sql.DB
	TicketRepo port.TicketRepository
}

func (u *CheckinTicket) Execute(token string) error {

	tx, err := u.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	ticket, err := u.TicketRepo.FindByTokenForUpdate(tx, token)
	if err != nil {
		return err
	}

	if ticket.CheckedInAt != nil {
		return ErrTicketAlreadyCheckedIn
	}

	err = u.TicketRepo.CheckIn(tx, ticket.Token)
	if err != nil {
		return err
	}

	return tx.Commit()
}