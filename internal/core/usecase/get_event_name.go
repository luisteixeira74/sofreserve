package usecase

import "sof-reserve/internal/core/port"

type GetEventNameUseCase struct {
	repo port.EventRepository
}

func NewGetEventNameUseCase(repo port.EventRepository) *GetEventNameUseCase {
	return &GetEventNameUseCase{repo: repo}
}

func (uc *GetEventNameUseCase) Execute(id int) (string, error) {
	event, err := uc.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	return event.Name, nil
}