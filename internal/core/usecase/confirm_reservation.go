package usecase

import (
	"database/sql"
	"errors"
	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
	"sof-reserve/internal/shared/security"
	"time"
)

type ConfirmReservationUseCase struct {
	db *sql.DB
}

type ConfirmReservationOutput struct {
	ReservationID int
	ReservationToken string
	Name     string
	Email    string
	Quantity int
	EventID  int
	Status   string
	Message  string
	Tickets []ReservationTicketOutput
}

type ReservationTicketOutput struct {
	Number    int
	Token	 string
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
	var reservationID int
	var reservationToken string

	err = tx.QueryRow(`
		SELECT id, name, email, quantity, event_id, status, token
		FROM reservations
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(
		&reservationID,
		&name,
		&email,
		&quantity,
		&eventID,
		&status,
		&reservationToken,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, coreErr.ErrReservationNotFound
		}
		return nil, err
	}

	var tickets []ReservationTicketOutput

	tickets, err = uc.loadReservationTickets(
		tx,
		reservationID,
	)

	if err != nil {
		return nil, err
	}

	// =====================
	// IDEMPOTÊNCIA
	// =====================
	if status == string(entity.StatusConfirmed) {
		_ = tx.Commit()

		return &ConfirmReservationOutput{
			ReservationID: reservationID,
			ReservationToken: reservationToken,
			Message:  "Reserva já confirmada.",
			Name:     name,
			Email:    email,
			Quantity: quantity,
			EventID:  eventID,
			Status:   status,
			Tickets:  tickets,
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

tickets, err = uc.generateTickets(
	tx,
	reservationID,
	eventID,
	quantity,
)

if err != nil {
	return nil, err
}

if err := tx.Commit(); err != nil {
	return nil, err
}

return &ConfirmReservationOutput{
	ReservationID:    reservationID,
	ReservationToken: reservationToken,
	Message:          "Reserva confirmada com sucesso!",
	Name:             name,
	Email:            email,
	Quantity:         quantity,
	EventID:          eventID,
	Status:           string(entity.StatusConfirmed),
	Tickets:          tickets,
}, nil}


func (uc *ConfirmReservationUseCase) loadReservationTickets(
	tx *sql.Tx,
	reservationID int,
) ([]ReservationTicketOutput, error) {

	rows, err := tx.Query(`
		SELECT
			ticket_number,
			token
		FROM reservation_tickets
		WHERE reservation_id = $1
		ORDER BY ticket_number
	`, reservationID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tickets []ReservationTicketOutput

	for rows.Next() {
		var ticket ReservationTicketOutput

		err := rows.Scan(
			&ticket.Number,
			&ticket.Token,
		)

		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (uc *ConfirmReservationUseCase) generateTickets(
	tx *sql.Tx,
	reservationID int,
	eventID int,
	quantity int,
) ([]ReservationTicketOutput, error) {

	var tickets []ReservationTicketOutput

	for i := 1; i <= quantity; i++ {

		ticketToken, err := security.GenerateTicketToken()
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(`
			INSERT INTO reservation_tickets (
				reservation_id,
				event_id,
				ticket_number,
				token
			)
			VALUES ($1, $2, $3, $4)
		`,
			reservationID,
			eventID,
			i,
			ticketToken,
		)

		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ReservationTicketOutput{
			Number: i,
			Token:  ticketToken,
		})
	}

	return tickets, nil
}