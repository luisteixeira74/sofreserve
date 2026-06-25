package port

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type EventRepository interface {
	GetByID(id int64) (entity.Event, error)

	GetByPublicID(publicID string) (entity.Event, error)

	FindByIDForUpdate(tx *sql.Tx, id int64) (entity.Event, error)

	FindByOwnerToken(token string) (entity.Event, error)

	CountEventsByOrganizerEmail(email string) (int64, error)

	Create(event entity.Event) (int64, error)
}