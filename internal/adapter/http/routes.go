package http

import (
	"database/sql"
	"net/http"

	"sof-reserve/internal/core/usecase"
)

func NewRouter(
	reserveUC *usecase.ReserveSpotUseCase,
	db *sql.DB,
) http.Handler {

	handler := &Handler{
		db:        db,
		reserveUC: reserveUC,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthHandler)
	mux.HandleFunc("/create-event", handler.CreateEventPage)
	mux.HandleFunc("/events", handler.CreateEventHandler)
	mux.HandleFunc("/evento", handler.EventPage)
	mux.HandleFunc("/reservations", handler.CreateReservationHandler)

	return mux
}