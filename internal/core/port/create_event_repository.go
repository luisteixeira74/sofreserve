package port

import "sof-reserve/internal/core/entity"

type CreateEventRepository interface {
	Create(event entity.Event) (int64, error)
}