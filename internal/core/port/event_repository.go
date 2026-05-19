package port

import (
	"database/sql"
	"sof-reserve/internal/core/entity"
)

type EventRepository interface {
	GetByID(id int) (entity.Event, error)
	GetByPublicID(publicID string) (entity.Event, error)
	FindByIDForUpdate(tx *sql.Tx, id int) (entity.Event, error)
	FindByOwnerToken(token string) (entity.Event, error)
	CountEventsByOrganizerEmail(email string) (int, error)
	Create(event entity.Event) (int, error)
}