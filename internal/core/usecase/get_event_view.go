package usecase

import (
	"fmt"
	"sof-reserve/internal/core/port"
	"time"
)

//
// =====================
// CLOCK (TESTABLE TIME)
// =====================
//

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

//
// =====================
// VIEW MODEL
// =====================
//

type EventView struct {
	ID            int64
	Name          string
	TotalSeats    int
	Reserved      int
	Available     int
	Percentage    int
	RemainingText string
	ShowAlert     bool
	IsClosed      bool
	PublicID      string
	OrganizerEmail string
}

//
// =====================
// USECASE
// =====================
//

type GetEventViewUseCase struct {
	eventRepo port.EventRepository
	resRepo   port.ReservationRepository
	clock     Clock
}

func NewGetEventViewUseCase(
	eventRepo port.EventRepository,
	resRepo port.ReservationRepository,
	clock Clock,
) *GetEventViewUseCase {
	return &GetEventViewUseCase{
		eventRepo: eventRepo,
		resRepo:   resRepo,
		clock:     clock,
	}
}

//
// =====================
// PUBLIC METHODS
// =====================
//

func (uc *GetEventViewUseCase) Execute(eventID int64) (EventView, error) {
	event, err := uc.eventRepo.GetByID(eventID)
	if err != nil {
		return EventView{}, err
	}

	reserved, err := uc.resRepo.SumByEventID(eventID)
	if err != nil {
		return EventView{}, err
	}

	return uc.build(
		event.ID,
		event.Name,
		event.TotalSeats,
		reserved,
		event.EndsAt,
		event.PublicID,
		event.OrganizerEmail,
	), nil
}

func (uc *GetEventViewUseCase) ExecuteByPublicID(publicID string) (EventView, error) {
	event, err := uc.eventRepo.GetByPublicID(publicID)
	if err != nil {
		return EventView{}, err
	}

	reserved, err := uc.resRepo.SumByEventID(event.ID)
	if err != nil {
		return EventView{}, err
	}

	return uc.build(
		event.ID,
		event.Name,
		event.TotalSeats,
		reserved,
		event.EndsAt,
		event.PublicID,
		event.OrganizerEmail,
	), nil
}

//
// =====================
// CORE BUILDER (DEDUP)
// =====================
//

func (uc *GetEventViewUseCase) build(
	id int64,
	name string,
	totalSeats int,
	reserved int,
	endsAt time.Time,
	publicID string,
	organizerEmail string,
) EventView {

	now := uc.clock.Now()
	remaining := endsAt.Sub(now)

	isClosed := remaining <= 0

	remainingText := ""
	showAlert := false

	if isClosed {
		remainingText = "Evento encerrado"
	} else {
		hours := int(remaining.Hours())

		if hours <= 48 {
			showAlert = true
		}

		days := hours / 24

		if days > 0 {
			remainingText = fmt.Sprintf("%d dias restantes", days)
		} else {
			remainingText = fmt.Sprintf("%d horas restantes", hours)
		}
	}

	percentage := 0
	if totalSeats > 0 {
		percentage = (reserved * 100) / totalSeats
	}

	return EventView{
		ID:            id,
		Name:          name,
		TotalSeats:    totalSeats,
		Reserved:      reserved,
		Available:     totalSeats - reserved,
		Percentage:    percentage,
		RemainingText: remainingText,
		ShowAlert:     showAlert,
		IsClosed:      isClosed,
		PublicID:      publicID,
		OrganizerEmail: organizerEmail,
	}
}