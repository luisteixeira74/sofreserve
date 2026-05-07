package http

import (
	"database/sql"
	"net/http"

	"sof-reserve/internal/core/port"
	"sof-reserve/internal/core/usecase"
)

func NewRouter(
	createReservationUC *usecase.CreateReservationUseCase,
	confirmUC *usecase.ConfirmReservationUseCase,
	eventViewUC *usecase.GetEventViewUseCase,
	eventRepo port.EventRepository,
	reservationRepo port.ReservationRepository,
	db *sql.DB,
) http.Handler {

	handler := &Handler{
		reserveUC:       createReservationUC,
		confirmUC:       confirmUC,
		eventViewUC:     eventViewUC,
		eventRepo:       eventRepo,
		reservationRepo: reservationRepo,
		db:              db,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthHandler)

	// pages
	mux.HandleFunc("/onboarding", handler.OnboardingPage)

	// events
	mux.HandleFunc("/events/new", handler.CreateEventPage)
	mux.HandleFunc("/events", handler.CreateEventHandler)

	// dashboard do evento
	mux.HandleFunc("/events/", handler.EventPageByPublicID)

	// evento público
	mux.HandleFunc("/e/", handler.EventPublicPage)

	// owner dashboard
	mux.HandleFunc("/manage/", handler.OwnerDashboard)

	// reservations
	mux.HandleFunc("/events/reserve", handler.CreateReservationHandler)

	// confirm / cancel
	mux.HandleFunc("/confirm", handler.ConfirmReservation)
	mux.HandleFunc("/cancel", handler.CancelReservation)

	// assets
	fs := http.FileServer(http.Dir("./internal/view/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	return mux
}