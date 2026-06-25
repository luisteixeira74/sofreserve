package usecase

import (
	"database/sql"

	coreErr "sof-reserve/internal/core/errors"
	"sof-reserve/internal/core/port"
)

type CancelReservationUseCase struct {
	db              *sql.DB
	reservationRepo port.ReservationRepository
}

func NewCancelReservationUseCase(
	db *sql.DB,
	reservationRepo port.ReservationRepository,
) *CancelReservationUseCase {
	return &CancelReservationUseCase{
		db:              db,
		reservationRepo: reservationRepo,
	}
}

func (uc *CancelReservationUseCase) Execute(token string) error {
	if token == "" {
		return coreErr.ErrInvalidToken
	}

	tx, err := uc.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	res, err := uc.reservationRepo.CancelIfAllowed(tx, token)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return coreErr.ErrInvalidStatusTransition
	}

	return tx.Commit()
}