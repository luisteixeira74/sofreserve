package port

import (
	"sof-reserve/internal/core/entity"
	"time"
)

type EventRepository interface {
	Create(name string, totalSeats int, endsAt time.Time) (int, error)

	GetByID(id int) (entity.Event, error)

	FindByIDForUpdate(id int) (int, time.Time, error)
}