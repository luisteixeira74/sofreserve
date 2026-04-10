package usecase

import (
	"errors"
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

func (uc *CreateEventUseCase) Execute(name string, totalSeats int, endsAt time.Time) (int, error) {
	if name == "" {
		return 0, errors.New("nome é obrigatório")
	}

	if totalSeats <= 0 {
		return 0, errors.New("total de vagas deve ser maior que zero")
	}

	if time.Now().After(endsAt) {
		return 0, errors.New("data de encerramento deve ser futura")
	}

	return uc.repo.Create(name, totalSeats, endsAt)
}