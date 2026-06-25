package usecase

import "sof-reserve/internal/core/port"

type GetOrganizerStatsUseCase struct {
    eventRepo port.EventRepository
}

func NewGetOrganizerStatsUseCase(repo port.EventRepository) *GetOrganizerStatsUseCase {
    return &GetOrganizerStatsUseCase{
        eventRepo: repo,
    }
}

type OrganizerStats struct {
    EventCount int64
}

func (uc *GetOrganizerStatsUseCase) Execute(email string) (OrganizerStats, error) {
    count, err := uc.eventRepo.CountEventsByOrganizerEmail(email)
    if err != nil {
        return OrganizerStats{}, err
    }

    return OrganizerStats{
        EventCount: count,
    }, nil
}