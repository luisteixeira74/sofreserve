package http

import (
	"database/sql"
	"net/http"

	"sof-reserve/internal/core/usecase"
)

func NewRouter(
	createReservationUC *usecase.CreateReservationUseCase,
	confirmUC *usecase.ConfirmReservationUseCase,
	eventViewUC *usecase.GetEventViewUseCase,
	db *sql.DB,
) http.Handler {

	handler := &Handler{
		reserveUC:   createReservationUC,
		confirmUC:   confirmUC,
		eventViewUC: eventViewUC,
		db:          db,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthHandler)

	// pages
	mux.HandleFunc("/onboarding", handler.OnboardingPage)

	// events
	mux.HandleFunc("/events/new", handler.CreateEventPage)
	mux.HandleFunc("/events", handler.CreateEventHandler) // POST
	// event dashboard
	mux.HandleFunc("/events/", handler.EventPageByPublicID) // GET /events/{public_id}


	// public event
	mux.HandleFunc("/e/", handler.EventPublicPage)

	// reservations
	mux.HandleFunc("/events/reserve", handler.CreateReservationHandler)

	// confirm / cancel
	mux.HandleFunc("/confirm", handler.ConfirmReservation)
	mux.HandleFunc("/cancel", handler.CancelReservation)

	// assets (filesystem)
	fs := http.FileServer(http.Dir("./internal/view/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	return mux
}