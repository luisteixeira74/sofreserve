package usecase

import (
	"database/sql"
	"errors"
	"fmt"
	"sof-reserve/internal/adapter/repository/postgres"
	"sof-reserve/internal/core/dto"
	coreErr "sof-reserve/internal/core/errors"
	"time"
)

type ReserveSpotUseCase struct {
	eventRepo       *postgres.EventRepository
	reservationRepo *postgres.ReservationRepository
}

func NewReserveSpotUseCase(
	eventRepo *postgres.EventRepository,
	reservationRepo *postgres.ReservationRepository,
) *ReserveSpotUseCase {
	return &ReserveSpotUseCase{
		eventRepo:       eventRepo,
		reservationRepo: reservationRepo,
	}
}

func (uc *ReserveSpotUseCase) Execute(req dto.ReserveRequest) error {

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

	// =====================
	// TRANSAÇÃO
	// =====================
	tx, err := uc.eventRepo.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// =====================
	// REGRA: EMAIL ÚNICO (apenas confirmed)
	// =====================
	exists, err := uc.reservationRepo.ExistsByEventAndEmail(tx, req.EventID, req.Email)
	if err != nil {
		return err
	}

	if exists {
		return coreErr.ErrEmailAlreadyUsed
	}

	// =====================
	// LOCK DO EVENTO
	// =====================
	var totalSeats int
	var endsAt time.Time

	err = tx.QueryRow(
		"SELECT total_seats, ends_at FROM events WHERE id = $1 FOR UPDATE",
		req.EventID,
	).Scan(&totalSeats, &endsAt)	

	// valida se o evento existe
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return coreErr.ErrEventNotFound
		}
		return err
	}

	// valida aqui se o evento já terminou
	if time.Now().After(endsAt) {
		return coreErr.ErrEventClosed
	}

	// =====================
	// SOMA APENAS CONFIRMED
	// =====================
	var totalReserved int
	err = tx.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) 
		 FROM reservations 
		 WHERE event_id = $1 AND status = 'confirmed'`,
		req.EventID,
	).Scan(&totalReserved)
	if err != nil {
		return err
	}

	available := totalSeats - totalReserved

	// =====================
	// REGRA: SEM OVERBOOKING
	// =====================
	if req.Quantity > available {
		return coreErr.ErrNotEnoughSeats
	}

	// =====================
	// GERAR TOKEN + STATUS
	// =====================
	token := generateToken(req.Email)
	status := "pending"

	// =====================
	// INSERT 
	// =====================
	err = uc.reservationRepo.Create(
		tx,
		req.EventID,
		req.Name,
		req.Email,
		req.Quantity,
		status,
		token,
	)
	if err != nil {
		return err
	}

	// =====================
	// "ENVIO" DE EMAIL (LOG)
	// =====================
	link := "http://localhost:8080/confirm?token=" + token
	fmt.Println("CONFIRM LINK:", link)

	// =====================
	// COMMIT
	// =====================
	return tx.Commit()
}

func generateToken(email string) string {
	token := fmt.Sprintf("%d-%s", time.Now().UnixNano(), email)
	return token
}
