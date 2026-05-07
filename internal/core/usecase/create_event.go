package usecase

import (
	coreErr "sof-reserve/internal/core/errors"
	"time"
)

type EventRepository interface {
	Create(name string, totalSeats int, endsAt time.Time) (int, error)
}

type CreateEventUseCase struct {
	repo EventRepository
}

func NewCreateEventUseCase(repo EventRepository) *CreateEventUseCase {
	return &CreateEventUseCase{repo: repo}
}

type CreateEventInput struct {
	Name       string
	TotalSeats int
	EndsAt     time.Time
}

func (uc *CreateEventUseCase) Execute(input CreateEventInput) (int, error) {

	if input.Name == "" {
		return 0, coreErr.ErrInvalidName
	}

	if input.TotalSeats <= 0 {
		return 0, coreErr.ErrInvalidQuantity
	}

	if time.Now().After(input.EndsAt) {
		return 0, coreErr.ErrEventClosed
	}

	return uc.repo.Create(
		input.Name,
		input.TotalSeats,
		input.EndsAt,
	)
}