package usecase

import (
	"database/sql"
	"fmt"
	"sof-reserve/internal/adapter/repository/postgres"
	"sof-reserve/internal/core/dto"
	"sof-reserve/internal/core/errors"
)

type ReserveSpotUseCase struct {
    eventRepo *postgres.EventRepository
    reservationRepo *postgres.ReservationRepository
}

func NewReserveSpotUseCase(
    db *sql.DB,
    eventRepo *postgres.EventRepository,
    reservationRepo *postgres.ReservationRepository,
) *ReserveSpotUseCase {
    return &ReserveSpotUseCase{
        eventRepo: eventRepo,
        reservationRepo: reservationRepo,
    }
}

func (uc *ReserveSpotUseCase) Execute(req dto.ReserveRequest) error {
	if req.EventID <= 0 {
    	return errors.ErrInvalidEventID
	}

	if req.Name == "" {
		return errors.ErrInvalidName
	}

	if req.Quantity <= 0 {
		return errors.ErrInvalidQuantity
	}
	
	tx, err := uc.eventRepo.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var totalSeats int

	err = tx.QueryRow(
		"SELECT total_seats FROM events WHERE id = $1 FOR UPDATE",
		req.EventID,
	).Scan(&totalSeats)
	if err != nil {
		return err
	}

	var totalReserved int

	err = tx.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM reservations WHERE event_id = $1",
		req.EventID,
	).Scan(&totalReserved)
	if err != nil {
		return err
	}

	available := totalSeats - totalReserved

	if req.Quantity > available {
		return errors.ErrNotEnoughSeats
	}

	_, err = tx.Exec(
		"INSERT INTO reservations (event_id, name, quantity) VALUES ($1, $2, $3)",
		req.EventID, req.Name, req.Quantity,
	)
	if err != nil {
		return err
	}

	fmt.Println("ENTERED USECASE")
	fmt.Println("totalSeats:", totalSeats)
	fmt.Println("totalReserved:", totalReserved)
	fmt.Println("qty:", req.Quantity)

	return tx.Commit()
}