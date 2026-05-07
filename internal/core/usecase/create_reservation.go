package usecase

import (
	"database/sql"
	"errors"

	"sof-reserve/internal/core/dto"
	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
	"sof-reserve/internal/core/port"
	"sof-reserve/internal/shared/security"
)

type CreateReservationUseCase struct {
	db              *sql.DB
	eventRepo       port.EventRepository
	reservationRepo port.ReservationRepository
	clock           Clock
}

func NewCreateReservationUseCase(
	db *sql.DB,
	eventRepo port.EventRepository,
	reservationRepo port.ReservationRepository,
	clock Clock,
) *CreateReservationUseCase {
	return &CreateReservationUseCase{
		db:              db,
		eventRepo:       eventRepo,
		reservationRepo: reservationRepo,
		clock:           clock,
	}
}

func (uc *CreateReservationUseCase) Execute(req dto.ReserveRequest) (string, error) {

	if req.EventID <= 0 {
		return "", coreErr.ErrInvalidEventID
	}
	if req.Name == "" {
		return "", coreErr.ErrInvalidName
	}
	if req.Email == "" {
		return "", coreErr.ErrInvalidEmail
	}
	if req.Quantity <= 0 {
		return "", coreErr.ErrInvalidQuantity
	}

	tx, err := uc.db.Begin()
	if err != nil {
		return "", err
	}

	defer tx.Rollback()

	event, err := uc.eventRepo.FindByIDForUpdate(tx, req.EventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", coreErr.ErrEventNotFound
		}
		return "", err
	}

	if uc.clock.Now().After(event.EndsAt) {
		return "", coreErr.ErrEventClosed
	}

	exists, err := uc.reservationRepo.ExistsByEventAndEmail(tx, req.EventID, req.Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", coreErr.ErrEmailAlreadyUsed
	}

	totalReserved, err := uc.reservationRepo.SumByEventID(req.EventID)
	if err != nil {
		return "", err
	}

	available := event.TotalSeats - totalReserved

	if req.Quantity > available {
		return "", coreErr.ErrNotEnoughSeats
	}

	token, err := security.GenerateToken()
	if err != nil {
		return "", err
	}

	err = uc.reservationRepo.Create(
		tx,
		req.EventID,
		req.Name,
		req.Email,
		req.Quantity,
		string(entity.StatusPending),
		token,
	)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return token, nil
}