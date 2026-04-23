package http

import (
	"database/sql"
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
	mux.HandleFunc("/onboarding", handler.OnboardingPage)

	// eventos (admin)
	mux.HandleFunc("/events/new", handler.CreateEventPage)   // GET
	mux.HandleFunc("/events/view", handler.EventPage)        // GET
	mux.HandleFunc("/events", handler.CreateEventHandler)    // POST

	mux.HandleFunc("/e/", handler.EventPublicPage)

	// reservas (guest)
	mux.HandleFunc("/reservation", handler.ReservationPage)      // GET
	mux.HandleFunc("/events/reserve", handler.CreateReservationHandler) // POST
	// confirmação
	mux.HandleFunc("/confirm", handler.ConfirmReservation)
	// cancelamento
	mux.HandleFunc("/cancel", handler.CancelReservation)

	// static files
	fs := http.FileServer(http.Dir("./templates/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	return mux
}