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

func NewConfirmReservationUseCase(db *sql.DB) *ConfirmReservationUseCase {
	return &ConfirmReservationUseCase{
		db: db,
	}
}

type ConfirmOutput struct {
	Name     string
	Email    string
	Quantity int
	EventID  int
	Status   string
	Message  string
	Token    string
}

func (uc *ConfirmReservationUseCase) Execute(token string) (*ConfirmOutput, error) {

	// =====================
	// VALIDAÇÃO BÁSICA
	// =====================
	if token == "" {
		return nil, coreErr.ErrInvalidToken
	}

	// =====================
	// TRANSAÇÃO
	// =====================
	tx, err := uc.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// =====================
	// BUSCA RESERVA
	// =====================
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
	// IDEMPOTÊNCIA
	// se a reserva já estiver confirmada, apenas retorna os dados sem tentar confirmar novamente
	// =====================
	if status != string(entity.StatusPending) {

		_ = tx.Commit()

		return &ConfirmOutput{
			Message:  "Reserva confirmada com sucesso!",
			Name:     name,
			Email:    email,
			Quantity: quantity,
			EventID:  eventID,
			Status:   string(entity.StatusConfirmed),
			Token:    token,
		}, nil
	}

	// =====================
	// VALIDA EVENTO
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
	// VERIFICA VAGAS
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
	// CONFIRMA RESERVA
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

	// =====================
	// COMMIT
	// =====================
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// =====================
	// OUTPUT
	// =====================
	return &ConfirmOutput{
		Message:  "Reserva confirmada com sucesso!",
		Name:     name,
		Email:    email,
		Quantity: quantity,
		EventID:  eventID,
		Status:   string(entity.StatusConfirmed),
		Token:    token,
	}, nil
}