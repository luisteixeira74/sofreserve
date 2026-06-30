package http

import (
	"database/sql"
	"net/http"
	"strings"

	"sof-reserve/internal/core/port"
	"sof-reserve/internal/core/usecase"
)

func NewRouter(
    createReservationUC *usecase.CreateReservationUseCase,
    confirmUC *usecase.ConfirmReservationUseCase,
    eventViewUC *usecase.GetEventViewUseCase,

    eventRepo port.EventRepository,
    reservationRepo port.ReservationRepository,
    ticketRepo port.TicketRepository,

    createEventUC *usecase.CreateEventUseCase,
    getOrganizerStatsUC *usecase.GetOrganizerStatsUseCase,
    checkinTicketUC *usecase.CheckinTicket,

    db *sql.DB,
) http.Handler {

	handler := &Handler{
		reserveUC:       createReservationUC,
		confirmUC:       confirmUC,
		eventViewUC:     eventViewUC,

		eventRepo:       eventRepo,
		reservationRepo: reservationRepo,
		ticketRepo:      ticketRepo,

		createEventUC:   createEventUC,
		organizerStatsUC:  getOrganizerStatsUC,
		checkinTicketUC: checkinTicketUC,

		db:              db,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthHandler)

	// pages
	mux.HandleFunc("/", handler.LandingPage)

	// events
	mux.HandleFunc("/events/new", handler.CreateEventPage) // GET
	mux.HandleFunc("/events", handler.CreateEventHandler) // POST

	// owner event list
	mux.HandleFunc("/manage/", handler.OwnerDashboard)

	// dashboard do evento
	mux.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {

			parts := strings.Split(
			strings.Trim(r.URL.Path, "/"),
			"/",
		)

		if len(parts) == 3 &&
			parts[2] == "reservations" {

			handler.EventReservationsPage(
				w,
				r,
				parts[1],
			)

			return
		}

		if len(parts) == 3 &&
			parts[2] == "checkin" {

			handler.OwnerCheckin(
				w,
				r,
			)

			return
		}

		handler.EventPageByPublicID(
			w,
			r,
		)
	})

	// link publico do evento
	mux.HandleFunc("/e/", handler.EventPublicPage)
	
	// reservation form
	mux.HandleFunc("/events/reserve", handler.CreateReservationHandler)
	
	// confirm / cancel
	mux.HandleFunc("/confirm", handler.ConfirmReservation)
	mux.HandleFunc("/cancel", handler.CancelReservation)

	mux.HandleFunc("/ticket/", handler.TicketView)

	// assets
	fs := http.FileServer(http.Dir("./internal/view/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	return mux
}