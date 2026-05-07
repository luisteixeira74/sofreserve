package usecase

import (
	"errors"
	"sof-reserve/internal/core/entity"
	"sof-reserve/internal/core/port"
)

type GetEventUseCase struct {
	repo port.EventRepository
}

func NewGetEventUseCase(repo port.EventRepository) *GetEventUseCase {
	return &GetEventUseCase{repo: repo}
}

func (uc *GetEventUseCase) Execute(id int) (entity.Event, error) {
	if id <= 0 {
		return entity.Event{}, errors.New("invalid event id")
	}

	return uc.repo.GetByID(id)
}