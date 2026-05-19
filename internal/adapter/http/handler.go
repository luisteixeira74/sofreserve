package http

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sof-reserve/internal/core/dto"
	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
	"sof-reserve/internal/core/port"
	"sof-reserve/internal/core/usecase"
	"sof-reserve/internal/shared/id"
	"sof-reserve/internal/shared/security"
)

var tmpl = template.Must(
	template.ParseGlob("internal/view/templates/*.html"),
)

// =====================
// VIEW MODELS
// =====================

type EventView struct {
	ID            int
	Name          string
	TotalSeats    int
	Reserved      int
	Available     int
	Percentage    int
	RemainingText string
	ShowAlert     bool
	IsClosed      bool
	PublicID      string
	LastUpdated   string
	PublicLink string
	OtherEvents string
	OrganizerEventCount int
	OrganizerEmail string
}

type ReservationPageData struct {
	Error       string
	EventID     int
	Name        string
	Email       string
	Quantity    int
	EventName   string
	Token       string
	ConfirmLink string
	Available   int
}

type RenderTemplateData struct {
	Page  string
	Title string
	Data  any
}

type ReservationConfirmedView struct {
	EventName string
	Name      string
	Email     string
	Quantity  int
	Token     string
	Message   string
	Status    string
	CancelLink string
}

type ReservationErrorView struct {
	Message string
	Status  string
}

type ReservationCancelView struct {
	Message   string
	EventID   int
	EventName string
	PublicLink string
}

type EventCreateView struct {
	Name           string
	OrganizerEmail string
	Error          string
}

type OwnerDashboardView struct {
	EventName     string
	TotalSeats    int
	TotalConfirmed int
	Reservations  []entity.Reservation
}

// =====================
// HANDLER
// =====================

type Handler struct {
	reserveUC   *usecase.CreateReservationUseCase
	confirmUC   *usecase.ConfirmReservationUseCase
	eventViewUC *usecase.GetEventViewUseCase
	createEventUC *usecase.CreateEventUseCase
	eventRepo port.EventRepository
	reservationRepo port.ReservationRepository
	organizerStats *usecase.GetOrganizerStats
	db *sql.DB
}

// =====================
// TEMPLATE
// =====================

func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data RenderTemplateData) {
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// =====================
// BASIC
// =====================

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func buildTitle(page string, context string) string {
	const appName = "SofReserve"

	if context != "" {
		return fmt.Sprintf("%s • %s • %s", context, page, appName)
	}
	return fmt.Sprintf("%s • %s", page, appName)
}

// =====================
// PAGES
// =====================

func (h *Handler) OnboardingPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "onboarding",
		Title: buildTitle("Onboarding", ""),
	})
}

func (h *Handler) CreateEventPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_create",
		Title: buildTitle("Criar evento", ""),
	})
}

// =====================
// EVENT CREATE
// =====================

func (h *Handler) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.ToLower(strings.TrimSpace(r.FormValue("organizer_email")))
	totalSeatsStr := r.FormValue("total_seats")

	totalSeats, err := strconv.Atoi(totalSeatsStr)

	if name == "" ||
		email == "" ||
		!strings.Contains(email, "@") ||
		err != nil ||
		totalSeats <= 0 {

		h.renderTemplate(w, "layout", RenderTemplateData{
			Page:  "event_create",
			Title: buildTitle("Criar evento", ""),
			Data: EventCreateView{
				Name:           name,
				OrganizerEmail: email,
				Error:          "Dados inválidos",
			},
		})
		return
	}

	endsAt, err := time.Parse("2006-01-02", r.FormValue("ends_at"))
	if err != nil {

		h.renderTemplate(w, "layout", RenderTemplateData{
			Page:  "event_create",
			Title: buildTitle("Criar evento", ""),
			Data: EventCreateView{
				Name:           name,
				OrganizerEmail: email,
				Error:          "Data inválida",
			},
		})

		return
	}

	publicID := id.GeneratePublicID()

	ownerToken, err := security.GenerateToken()
	if err != nil {
		http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
		return
	}

	eventID, err := h.createEventUC.Execute(
		usecase.CreateEventInput{
			Name:           name,
			TotalSeats:     totalSeats,
			EndsAt:         endsAt,
			PublicID:       publicID,
			OrganizerEmail: email,
			OwnerToken:     ownerToken,
		},
	)

	if err != nil {

		h.renderTemplate(w, "layout", RenderTemplateData{
			Page:  "event_create",
			Title: buildTitle("Criar evento", ""),
			Data: EventCreateView{
				Name:           name,
				OrganizerEmail: email,
				Error:          err.Error(),
			},
		})

		return
	}

	_ = eventID

	http.Redirect(w, r, "/events/"+publicID, http.StatusSeeOther)
}

// =====================
// EVENT VIEW
// =====================

func (h *Handler) EventPageByPublicID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) < 2 || parts[1] == "" {
		http.NotFound(w, r)
		return
	}

	publicID := parts[1]

	// reservations page
	if len(parts) == 3 && parts[2] == "reservations" {
		h.EventReservationsPage(w, r, publicID)
		return
	}

	ucView, err := h.eventViewUC.ExecuteByPublicID(publicID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	baseURL := getBaseURL(r)

	// DEFENSIVO (garante consistência)
	available := ucView.TotalSeats - ucView.Reserved
	if available < 0 {
		available = 0
	}

	percentage := 0
	if ucView.TotalSeats > 0 {
		percentage = (ucView.Reserved * 100) / ucView.TotalSeats
	}

	stats, err := h.organizerStats.Execute(ucView.OrganizerEmail)
	if err != nil {
		stats.EventCount = 0
	}

	view := EventView{
		ID:            ucView.ID,
		Name:          ucView.Name,
		TotalSeats:    ucView.TotalSeats,
		Reserved:      ucView.Reserved,
		Available:     available,
		Percentage:    percentage,
		RemainingText: ucView.RemainingText,
		ShowAlert:     ucView.ShowAlert,
		IsClosed:      ucView.IsClosed,
		PublicID:      ucView.PublicID,
		PublicLink:    baseURL + "/e/" + ucView.PublicID,
		OrganizerEmail: ucView.OrganizerEmail,
		OrganizerEventCount: stats.EventCount,
		LastUpdated:   time.Now().Format("15:04:05"),
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_dashboard",
		Title: buildTitle("Dashboard", view.Name),
		Data:  view,
	})
}

func (h *Handler) EventReservationsPage(
	w http.ResponseWriter,
	r *http.Request,
	publicID string,
) {
	event, err := h.eventViewUC.ExecuteByPublicID(publicID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	reservations, err := h.reservationRepo.FindConfirmedByEventID(event.ID)
	if err != nil {
		http.Error(w, "failed to load reservations", http.StatusInternalServerError)
		return
	}

	data := struct {
		EventName     string
		PublicID      string
		TotalSeats    int
		TotalConfirmed int
		Reservations  []entity.Reservation
	}{
		EventName:      event.Name,
		PublicID:       event.PublicID,
		TotalSeats:     event.TotalSeats,
		TotalConfirmed: event.Reserved,
		Reservations:   reservations,
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_owner_dashboard",
		Title: buildTitle("Reservas", event.Name),
		Data:  data,
	})
}

// =====================
// PUBLIC EVENT
// =====================

func (h *Handler) EventPublicPage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	publicID := parts[1]

	// /e/{id}/reserve
	if len(parts) == 3 && parts[2] == "reserve" {
		h.reservationPageByPublicID(w, r, publicID)
		return
	}

	// /e/{id}
	ucView, err := h.eventViewUC.ExecuteByPublicID(publicID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	baseURL := getBaseURL(r)

	view := EventView{
		ID:            ucView.ID,
		Name:          ucView.Name,
		TotalSeats:    ucView.TotalSeats,
		Reserved:      ucView.Reserved,
		Available:     ucView.Available,
		Percentage:    ucView.Percentage,
		RemainingText: ucView.RemainingText,
		ShowAlert:     ucView.ShowAlert,
		IsClosed:      ucView.IsClosed,
		PublicID:      ucView.PublicID,
		PublicLink:    baseURL + "/e/" + ucView.PublicID,
		LastUpdated:   time.Now().Format("15:04:05"),
		OtherEvents: "Em breve: outros eventos do organizador", // Placeholder, pode ser implementado futuramente
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_public",
		Title: buildTitle("Reservas", view.Name),
		Data:  view,
	})
}

func (h *Handler) reservationPageByPublicID(w http.ResponseWriter, r *http.Request, publicID string) {
	view, err := h.eventViewUC.ExecuteByPublicID(publicID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	available := view.TotalSeats - view.Reserved
	if available < 0 {
		available = 0
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "reservation_form",
		Title: buildTitle("Reservar", view.Name),
		Data: ReservationPageData{
			EventID:   view.ID,
			EventName: view.Name,
			Available:  available, // 👈 ESSENCIAL
		},
	})
}

// =====================
// RESERVATION CREATE (POST)
// =====================

func (h *Handler) CreateReservationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// =====================
	// 1. PARSE + VALIDATION (INPUT)
	// =====================

	eventID, qty, name, email, err := h.parseReservationInput(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// =====================
	// 2. LOAD EVENT (SOURCE OF TRUTH)
	// =====================

	view, err := h.eventViewUC.Execute(eventID)
	if err != nil {
		http.Error(w, "evento não encontrado", http.StatusNotFound)
		return
	}

	// =====================
	// 3. BUSINESS RULE (AVAILABILITY CHECK)
	// =====================

	if qty > view.Available {
		h.renderReservation(w, ReservationPageData{
			EventID:   eventID,
			Name:      name,
			Email:     email,
			Quantity:  qty,
			EventName: view.Name,
			Error: fmt.Sprintf(
				"Você tentou reservar %d vagas, mas existem apenas %d disponíveis",
				qty,
				view.Available,
			),
			Available: view.Available,
		})
		return
	}

	// =====================
	// 4. EXECUTE USECASE
	// =====================

	token, err := h.reserveUC.Execute(dto.ReserveRequest{
		EventID:   eventID,
		EventName: view.Name,
		Name:      name,
		Email:     email,
		Quantity:  qty,
	})

	if err != nil {
		h.handleReservationError(w, err, ReservationPageData{
			EventID:   eventID,
			Name:      name,
			Email:     email,
			Quantity:  qty,
			EventName: view.Name,
			Available: view.Available,
		})
		return
	}

	// =====================
	// 5. SUCCESS FLOW
	// =====================

	baseURL := getBaseURL(r)

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "reservation_pending",
		Title: buildTitle("Reserva pendente", view.Name),
		Data: ReservationPageData{
			EventID:     eventID,
			Name:        name,
			Email:       email,
			Quantity:    qty,
			EventName:   view.Name,
			Token:       token,
			ConfirmLink: baseURL + "/confirm?token=" + token,
		},
	})
}

func (h *Handler) renderReservation(w http.ResponseWriter, data ReservationPageData) {
	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "reservation_form",
		Title: buildTitle("Reservar", data.EventName),
		Data:  data,
	})
}

// =====================
// CONFIRM
// =====================

func (h *Handler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		http.Error(w, "token inválido", http.StatusBadRequest)
		return
	}

	output, err := h.confirmUC.Execute(token)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	var eventName string

	err = h.db.QueryRow(
		"SELECT name FROM events WHERE id=$1",
		output.EventID,
	).Scan(&eventName)

	if err != nil {
		eventName = "Evento"
	}

	baseURL := getBaseURL(r)

	cancelLink := baseURL + "/cancel?token=" + output.Token

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "reservation_confirmed",
		Title: buildTitle("Confirmado", ""),
		Data: ReservationConfirmedView{
			EventName: eventName,
			Name:      output.Name,
			Email:     output.Email,
			Quantity:  output.Quantity,
			Token:     output.Token,
			Message:   output.Message,
			Status:    output.Status,
			CancelLink: cancelLink,
		},
	})
}

// =====================
// CANCEL
// =====================

func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token inválido", http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var view ReservationCancelView
	var status string

	err = tx.QueryRow(`
		SELECT
			r.event_id,
			r.status,
			e.name,
			e.public_link
		FROM reservations r
		JOIN events e ON e.id = r.event_id
		WHERE r.token = $1
		FOR UPDATE
	`, token).Scan(
		&view.EventID,
		&status,
		&view.EventName,
		&view.PublicLink,
	)

	if err != nil {
		http.Error(w, "não encontrado", http.StatusNotFound)
		return
	}

	if entity.ReservationStatus(status) == entity.StatusCanceled {
		view.Message = "Já cancelada"

		_ = tx.Commit()

		h.renderTemplate(w, "layout", RenderTemplateData{
			Page:  "reservation_cancel",
			Title: buildTitle("Cancelado", ""),
			Data:  view,
		})

		return
	}

	_, err = tx.Exec(`
		UPDATE reservations
		SET status = $1
		WHERE token = $2
	`, string(entity.StatusCanceled), token)

	if err != nil {
		http.Error(w, "erro ao cancelar reserva", http.StatusInternalServerError)
		return
	}

	view.Message = "Cancelado com sucesso"

	if err := tx.Commit(); err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "reservation_cancel",
		Title: buildTitle("Cancelado", ""),
		Data:  view,
	})
}

func (h *Handler) OwnerDashboard(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	token := parts[1]

	event, err := h.eventRepo.FindByOwnerToken(token)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	reservations, err := h.reservationRepo.FindConfirmedByEventID(event.ID)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	totalConfirmed := 0

	for _, reservation := range reservations {
		totalConfirmed += reservation.Quantity
	}

	view := OwnerDashboardView{
		EventName:      event.Name,
		TotalSeats:     event.TotalSeats,
		TotalConfirmed: totalConfirmed,
		Reservations:   reservations,
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_details_dashboard",
		Title: buildTitle("Owner Dashboard", event.Name),
		Data:  view,
	})
}

func (h *Handler) OwnerCheckin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func (h *Handler) handleReservationError(w http.ResponseWriter, err error, data ReservationPageData) {
	var appErr *coreErr.AppError

	if errors.As(err, &appErr) {
		data.Error = appErr.Message
		h.renderReservation(w, data)
		return
	}

	http.Error(w, "erro interno", http.StatusInternalServerError)
}

func (h *Handler) parseReservationInput(r *http.Request) (int, int, string, string, error) {
	eventID, err := strconv.Atoi(r.FormValue("event_id"))
	if err != nil || eventID <= 0 {
		return 0, 0, "", "", errors.New("evento inválido")
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || qty <= 0 {
		return 0, 0, "", "", errors.New("quantidade inválida")
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))

	if name == "" || email == "" || !strings.Contains(email, "@") {
		return 0, 0, "", "", errors.New("dados inválidos")
	}

	return eventID, qty, name, email, nil
}