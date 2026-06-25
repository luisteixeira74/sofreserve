package postgres

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(
	tx *sql.Tx,
	reservationID int64,
	eventID int64,
	ticketNumber int,
	token string,
) error {

	_, err := tx.Exec(`
		INSERT INTO reservation_tickets
		(reservation_id, event_id, ticket_number, token)
		VALUES ($1, $2, $3, $4)
	`,
		reservationID,
		eventID,
		ticketNumber,
		token,
	)

	return err
}

func (r *TicketRepository) FindByReservationID(
	reservationID int64,
) ([]entity.Ticket, error) {

	rows, err := r.db.Query(`
		SELECT
			id,
			reservation_id,
			ticket_number,
			token,
			checked_in_at,
			created_at
		FROM reservation_tickets
		WHERE reservation_id = $1
		ORDER BY ticket_number ASC
	`, reservationID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []entity.Ticket

	for rows.Next() {
		var ticket entity.Ticket

		err := rows.Scan(
			&ticket.ID,
			&ticket.ReservationID,
			&ticket.TicketNumber,
			&ticket.Token,
			&ticket.CheckedInAt,
			&ticket.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (r *TicketRepository) FindByToken(
	token string,
) (entity.Ticket, error) {

	var ticket entity.Ticket

	err := r.db.QueryRow(`
		SELECT
			id,
			reservation_id,
			ticket_number,
			token,
			checked_in_at,
			created_at
		FROM reservation_tickets
		WHERE token = $1
	`,
		token,
	).Scan(
		&ticket.ID,
		&ticket.ReservationID,
		&ticket.TicketNumber,
		&ticket.Token,
		&ticket.CheckedInAt,
		&ticket.CreatedAt,
	)

	return ticket, err
}

func (r *TicketRepository) FindByTokenForUpdate(
	tx *sql.Tx,
	token string,
) (entity.Ticket, error) {

	var ticket entity.Ticket

	err := tx.QueryRow(`
		SELECT
			id,
			reservation_id,
			ticket_number,
			token,
			checked_in_at,
			created_at
		FROM reservation_tickets
		WHERE token = $1
		FOR UPDATE
	`,
		token,
	).Scan(
		&ticket.ID,
		&ticket.ReservationID,
		&ticket.TicketNumber,
		&ticket.Token,
		&ticket.CheckedInAt,
		&ticket.CreatedAt,
	)

	return ticket, err
}

func (r *TicketRepository) CheckIn(
	tx *sql.Tx,
	token string,
) error {

	res, err := tx.Exec(`
		UPDATE reservation_tickets
		SET checked_in_at = NOW()
		WHERE token = $1
		AND checked_in_at IS NULL
	`,
		token,
	)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return coreErr.ErrTicketAlreadyCheckedIn
	}

	return nil
}

func (r *TicketRepository) FindTicketViewByToken(
	token string,
) (entity.TicketView, error) {

	var ticket entity.TicketView

	err := r.db.QueryRow(`
		SELECT
			e.name,
			t.token,
			t.ticket_number,
			t.checked_in_at
		FROM reservation_tickets t
		JOIN reservations r
			ON r.id = t.reservation_id
		JOIN events e
			ON e.id = r.event_id
		WHERE t.token = $1
	`,
		token,
	).Scan(
		&ticket.EventName,
		&ticket.Token,
		&ticket.TicketNumber,
		&ticket.CheckedInAt,
	)

	if err != nil {
		return entity.TicketView{}, err
	}

	return ticket, nil
}

func (r *TicketRepository) GetLastCheckinsByEventID(
	eventID int64,
	limit int,
) ([]entity.LastCheckin, error) {

	query := `
		SELECT 
			t.ticket_number,
			t.token,
			t.checked_in_at
		FROM reservation_tickets t
		WHERE t.event_id = $1
		  AND t.checked_in_at IS NOT NULL
		ORDER BY t.checked_in_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, eventID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entity.LastCheckin

	for rows.Next() {
		var lc entity.LastCheckin

		err := rows.Scan(
			&lc.TicketNumber,
			&lc.Token,
			&lc.CheckedInAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, lc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}