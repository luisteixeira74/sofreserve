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

	defer func() {
		_ = tx.Rollback()
	}()

	ticket, err := u.TicketRepo.FindByToken(token)
	if err == sql.ErrNoRows {
		return appErrors.ErrInvalidToken
	}
	if err != nil {
		return err
	}

	// 2. tentativa de check-in (estado real)
	rows, err := u.TicketRepo.MarkCheckinIfValid(tx, ticket.Token)
	if err != nil {
		return err
	}

	if rows == 0 {
		return appErrors.ErrTicketAlreadyCheckedIn
	}

	return tx.Commit()
}