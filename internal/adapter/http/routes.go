package http

import (
	"database/sql"
	"html/template"
	"net/http"

	"sof-reserve/internal/core/usecase"
)

func NewRouter(
	reserveUC *usecase.ReserveSpotUseCase,
	confirmUC *usecase.ConfirmReservationUseCase,
	db *sql.DB,
) http.Handler {

	handler := &Handler{
		reserveUC: reserveUC,
		confirmUC: confirmUC,
		db:        db,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthHandler)

	// onboarding
	mux.HandleFunc("/create-event", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/onboarding", http.StatusFound)
	})

	mux.HandleFunc("/onboarding", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles(
			"templates/layout.html",
			"templates/onboarding.html",
		))

		err := tmpl.ExecuteTemplate(w, "layout", nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
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

	// cancel reserva
	mux.HandleFunc("/cancel", handler.CancelReservation)

	// static files
	fs := http.FileServer(http.Dir("./templates/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	return mux
}