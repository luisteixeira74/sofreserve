package usecase

import (
	"time"

	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
	"sof-reserve/internal/core/port"
)

type CreateEventUseCase struct {
	repo port.EventRepository
}

func NewCreateEventUseCase(repo port.EventRepository) *CreateEventUseCase {
	return &CreateEventUseCase{
		repo: repo,
	}
}

type CreateEventInput struct {
	Name           string
	TotalSeats     int
	EndsAt         time.Time
	PublicID       string
	OrganizerEmail string
	OwnerToken     string
}

func (uc *CreateEventUseCase) Execute(input CreateEventInput) (int64, error) {

	if input.Name == "" {
		return 0, coreErr.ErrInvalidName
	}

	if input.TotalSeats <= 0 {
		return 0, coreErr.ErrInvalidQuantity
	}

	if time.Now().After(input.EndsAt) {
		return 0, coreErr.ErrEventClosed
	}

	event := entity.Event{
		Name:           input.Name,
		TotalSeats:     input.TotalSeats,
		EndsAt:         input.EndsAt,
		PublicID:       input.PublicID,
		OrganizerEmail: input.OrganizerEmail,
		OwnerToken:     input.OwnerToken,
	}

	return uc.repo.Create(event)
}