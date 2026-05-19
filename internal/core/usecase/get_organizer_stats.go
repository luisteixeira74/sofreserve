package usecase

import "sof-reserve/internal/core/port"

type GetOrganizerStats struct {
    eventRepo port.EventRepository
}

func NewGetOrganizerStatsUseCase(repo port.EventRepository) *GetOrganizerStats {
    return &GetOrganizerStats{
        eventRepo: repo,
    }
}

type OrganizerStats struct {
    EventCount int
}

func (uc *GetOrganizerStats) Execute(email string) (OrganizerStats, error) {
    count, err := uc.eventRepo.CountEventsByOrganizerEmail(email)
    if err != nil {
        return OrganizerStats{}, err
    }

    return OrganizerStats{
        EventCount: count,
    }, nil
}