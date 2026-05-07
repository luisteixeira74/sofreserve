package usecase

import (
	"database/sql"
	"errors"

	"sof-reserve/internal/core/entity"
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

	defer tx.Rollback()

	reservation, err := uc.reservationRepo.FindByTokenForUpdate(tx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return coreErr.ErrReservationNotFound
		}
		return err
	}

	currentStatus := entity.ReservationStatus(reservation.Status)

	if currentStatus == entity.StatusCanceled {
		return tx.Commit()
	}

	if !entity.CanTransition(currentStatus, entity.StatusCanceled) {
		return coreErr.ErrInvalidStatusTransition
	}

	err = uc.reservationRepo.UpdateStatus(
		tx,
		reservation.ID,
		string(entity.StatusCanceled),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}