package http

import (
	"database/sql"
	"html/template"
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

	// onboarding
	mux.HandleFunc("/create-event", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/onboarding", http.StatusFound)
	})

	mux.HandleFunc("/onboarding", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/onboarding.html"))
		tmpl.Execute(w, nil)
	})

	mux.HandleFunc("/create-event/form", handler.CreateEventPage)

	// eventos
	mux.HandleFunc("/events", handler.CreateEventHandler)
	mux.HandleFunc("/evento", handler.EventPage)

	// reservas
	mux.HandleFunc("/reservation", handler.ReservationPage)
	mux.HandleFunc("/events/reserve", handler.CreateReservationHandler)

	// confirmação
	mux.HandleFunc("/confirm", handler.ConfirmReservation)

	return mux
}