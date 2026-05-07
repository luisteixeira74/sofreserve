package usecase

import (
	"database/sql"
	"errors"
	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
	"time"
)

type ConfirmReservationUseCase struct {
	db *sql.DB
}

type ConfirmReservationOutput struct {
	Name     string
	Email    string
	Quantity int
	EventID  int
	Status   string
	Message  string
	Token    string
}

func NewConfirmReservationUseCase(db *sql.DB) *ConfirmReservationUseCase {
	return &ConfirmReservationUseCase{
		db: db,
	}
}

func (uc *ConfirmReservationUseCase) Execute(token string) (*ConfirmReservationOutput, error) {

	if token == "" {
		return nil, coreErr.ErrInvalidToken
	}

	tx, err := uc.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var name string
	var email string
	var quantity int
	var eventID int
	var status string

	err = tx.QueryRow(`
		SELECT name, email, quantity, event_id, status
		FROM reservations
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(&name, &email, &quantity, &eventID, &status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, coreErr.ErrReservationNotFound
		}
		return nil, err
	}

	// =====================
	// IDEMPOTÊNCIA CORRETA
	// =====================
	if status == string(entity.StatusConfirmed) {
		_ = tx.Commit()

		return &ConfirmReservationOutput{
			Message:  "Reserva já confirmada.",
			Name:     name,
			Email:    email,
			Quantity: quantity,
			EventID:  eventID,
			Status:   status,
			Token:    token,
		}, nil
	}

	// =====================
	// EVENTO
	// =====================
	var totalSeats int
	var endsAt time.Time

	err = tx.QueryRow(`
		SELECT total_seats, ends_at
		FROM events
		WHERE id = $1
		FOR UPDATE
	`, eventID).Scan(&totalSeats, &endsAt)

	if err != nil {
		return nil, coreErr.ErrEventNotFound
	}

	if time.Now().After(endsAt) {
		return nil, coreErr.ErrEventClosed
	}

	// =====================
	// CAPACIDADE
	// =====================
	var reserved int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(quantity), 0)
		FROM reservations
		WHERE event_id = $1 AND status = $2
	`, eventID, string(entity.StatusConfirmed)).Scan(&reserved)

	if err != nil {
		return nil, err
	}

	if reserved+quantity > totalSeats {
		return nil, coreErr.ErrNotEnoughSeats
	}

	// =====================
	// UPDATE
	// =====================
	_, err = tx.Exec(`
		UPDATE reservations
		SET status = $1
		WHERE token = $2
	`,
		string(entity.StatusConfirmed),
		token,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ConfirmReservationOutput{
		Message:  "Reserva confirmada com sucesso!",
		Name:     name,
		Email:    email,
		Quantity: quantity,
		EventID:  eventID,
		Status:   string(entity.StatusConfirmed),
		Token:    token,
	}, nil
}