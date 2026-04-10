package usecase

import (
	"errors"
	"fmt"
	"sof-reserve/internal/core/dto"
	coreErr "sof-reserve/internal/core/errors"
	"sof-reserve/internal/core/port"
	"time"
)

type ReserveEventUseCase struct {
	eventRepo       port.EventRepository
	reservationRepo port.ReservationRepository
}

func NewReserveEventUseCase(
	eventRepo port.EventRepository,
	reservationRepo port.ReservationRepository,
) *ReserveEventUseCase {
	return &ReserveEventUseCase{
		eventRepo:       eventRepo,
		reservationRepo: reservationRepo,
	}
}

func (uc *ReserveEventUseCase) Execute(req dto.ReserveRequest) error {

	// =====================
	// VALIDAÇÕES
	// =====================
	if req.EventID <= 0 {
		return coreErr.ErrInvalidEventID
	}
	if req.Name == "" {
		return coreErr.ErrInvalidName
	}
	if req.Email == "" {
		return coreErr.ErrInvalidEmail
	}
	if req.Quantity <= 0 {
		return coreErr.ErrInvalidQuantity
	}

	exists, err := uc.reservationRepo.ExistsByEventAndEmail(req.EventID, req.Email)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("reservation already exists")
	}

	total, err := uc.reservationRepo.SumConfirmedByEvent(req.EventID)
	if err != nil {
		return err
	}

	_ = total // (por enquanto não usado)

	token := generateToken(req.Email)

	err = uc.reservationRepo.Create(
		req.EventID,
		req.Name,
		req.Email,
		req.Quantity,
		"confirmed",
		token,
	)
	if err != nil {
		return err
	}

	return nil
}

func generateToken(email string) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), email)
}