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
	"sof-reserve/internal/core/usecase"
)

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
}

type Handler struct {
	db        *sql.DB
	reserveUC *usecase.ReserveSpotUseCase
	confirmUC *usecase.ConfirmReservationUseCase
}

type ReservationPageData struct {
	Error     string
	EventID   int
	Name      string
	Email     string
	Quantity  int
	EventName string
}

type RenderTemplateData struct {
	Page string
	Title string
	Data any
}

type ReservationConfirmedView struct {
	EventName string
	Name      string
	Email     string
	Quantity  int
	Token     string
	Message   string
	Status    string
}

type ReservationErrorView struct {
	Message string
	Status  string
}

type ReservationCancelView struct {
	Message string
	EventID int
}

type EventCreateView struct {
	Name  string
	Error string
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) OnboardingPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "onboarding",
		Title: buildTitle("Onboarding", ""),
	})
}

func (h *Handler) CreateEventPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "event_create",
		Title: buildTitle("Criar evento", ""),
	})
}

func (h *Handler) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	totalSeatsStr := r.FormValue("total_seats")

	totalSeats, err := strconv.Atoi(totalSeatsStr)
	if name == "" || err != nil || totalSeats <= 0 || totalSeats > 1000 {
		h.renderTemplate(w, "layout", RenderTemplateData{
			Page: "event_create",
			Title: buildTitle("Criar evento", ""),
			Data: EventCreateView{
				Name:  name,
				Error: "Dados inválidos",
			},
		})
		return
	}

	endsAt, err := time.Parse("2006-01-02", r.FormValue("ends_at"))
	if err != nil {
		h.renderTemplate(w, "layout", RenderTemplateData{
			Page: "event_create",
			Title: buildTitle("Criar evento", ""),
			Data: EventCreateView{
				Name:  name,
				Error: "Data inválida",
			},
		})
		return
	}

	publicID := strconv.FormatInt(time.Now().UnixNano(), 36)

	var id int
	err = h.db.QueryRow(
		"INSERT INTO events (name, total_seats, ends_at, public_id) VALUES ($1, $2, $3, $4) RETURNING id",
		name, totalSeats, endsAt, publicID,
	).Scan(&id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/events/view?id="+strconv.Itoa(id), http.StatusSeeOther)
}

//
// =====================
// CORE VIEW BUILDER
// =====================
//

func (h *Handler) buildEventView(id int, name string, totalSeats int, endsAt time.Time, publicID string) EventView {
	now := time.Now()
	remaining := endsAt.Sub(now)

	isClosed := remaining <= 0

	remainingText := ""
	showAlert := false

	if isClosed {
		remainingText = "Evento encerrado"
	} else {
		hours := int(remaining.Hours())

		if hours <= 48 {
			showAlert = true
		}

		days := hours / 24

		if days > 0 {
			remainingText = fmt.Sprintf("%d dias restantes", days)
		} else {
			remainingText = fmt.Sprintf("%d horas restantes", hours)
		}
	}

	var reserved int
	_ = h.db.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM reservations WHERE event_id = $1 AND status = 'confirmed'",
		id,
	).Scan(&reserved)

	percentage := 0
	if totalSeats > 0 {
		percentage = (reserved * 100) / totalSeats
	}

	return EventView{
		ID:            id,
		Name:          name,
		TotalSeats:    totalSeats,
		Reserved:      reserved,
		Available:     totalSeats - reserved,
		Percentage:    percentage,
		RemainingText: remainingText,
		ShowAlert:     showAlert,
		IsClosed:      isClosed,
		PublicID:      publicID,
	}
}

//
// =====================
// EVENT ADMIN
// =====================
//

func (h *Handler) EventPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	var name string
	var totalSeats int
	var endsAt time.Time
	var publicID string

	err = h.db.QueryRow(
		"SELECT name, total_seats, ends_at, public_id FROM events WHERE id = $1",
		id,
	).Scan(&name, &totalSeats, &endsAt, &publicID)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	view := h.buildEventView(id, name, totalSeats, endsAt, publicID)

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "event_dashboard",
		Title: buildTitle("Dashboard", name),
		Data: view,
	})
}

//
// =====================
// EVENT PUBLIC
// =====================
//

func (h *Handler) EventPublicPage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	publicID := parts[2]

	var id int
	var name string
	var totalSeats int
	var endsAt time.Time

	err := h.db.QueryRow(
		"SELECT id, name, total_seats, ends_at FROM events WHERE public_id = $1",
		publicID,
	).Scan(&id, &name, &totalSeats, &endsAt)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	view := h.buildEventView(id, name, totalSeats, endsAt, publicID)

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "event_public",
		Title: buildTitle("Reservas", ""),
		Data: view,
	})
}

//
// =====================
// RESERVATION
// =====================
//

func (h *Handler) CreateReservationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventID, errID := strconv.Atoi(r.FormValue("event_id"))
	qty, errQty := strconv.Atoi(r.FormValue("quantity"))

	data := ReservationPageData{
		EventID:  eventID,
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Quantity: qty,
	}

	if errID != nil || eventID <= 0 {
		data.Error = "Evento inválido"
		h.renderReservation(w, data)
		return
	}

	var eventName string
	err := h.db.QueryRow(
		"SELECT name FROM events WHERE id = $1",
		eventID,
	).Scan(&eventName)

	if err != nil {
		data.Error = "Evento não encontrado"
		h.renderReservation(w, data)
		return
	}
	data.EventName = eventName

	if errQty != nil || qty <= 0 || qty > 10 {
		data.Error = "Quantidade inválida"
		h.renderReservation(w, data)
		return
	}

	if data.Name == "" {
		data.Error = "Nome é obrigatório"
		h.renderReservation(w, data)
		return
	}

	if data.Email == "" || !strings.Contains(data.Email, "@") {
		data.Error = "Email inválido"
		h.renderReservation(w, data)
		return
	}

	err = h.reserveUC.Execute(dto.ReserveRequest{
		EventID:   eventID,
		EventName: eventName,
		Name:      data.Name,
		Email:     data.Email,
		Quantity:  qty,
	})

	if err != nil {
		var appErr *coreErr.AppError
		if errors.As(err, &appErr) {
			data.Error = appErr.Message
			h.renderReservation(w, data)
			return
		}
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "reservation_pending",
		Title: buildTitle("Reserva pendente", ""),
		Data: data,
	})
}

func (h *Handler) renderReservation(w http.ResponseWriter, data ReservationPageData) {
	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "reservation_form",
		Title: buildTitle("Reservar", data.EventName),
		Data: data,
	})
}

func (h *Handler) ReservationPage(w http.ResponseWriter, r *http.Request) {
	eventID, _ := strconv.Atoi(r.URL.Query().Get("event_id"))

	var eventName string
	err := h.db.QueryRow(
		"SELECT name FROM events WHERE id = $1",
		eventID,
	).Scan(&eventName)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.renderReservation(w, ReservationPageData{
		EventID:   eventID,
		EventName: eventName,
	})
}

//
// =====================
// CONFIRM
// =====================
//

func (h *Handler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	output, err := h.confirmUC.Execute(token)
	if err != nil {
		var appErr *coreErr.AppError
		if errors.As(err, &appErr) {
			h.renderTemplate(w, "layout", RenderTemplateData{
				Page: "reservation_error",
				Title: buildTitle("Erro na confirmação", ""),
				Data: ReservationErrorView{
					Message: appErr.Message,
					Status:  "error",
				},
			})
			return
		}
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	var eventName string

	err = h.db.QueryRow(
		"SELECT name FROM events WHERE id = $1",
		output.EventID,
	).Scan(&eventName)

	if err != nil {
		http.Error(w, "erro ao carregar evento", http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "reservation_confirmed",
		Title: "Reserva confirmada • Sof/Reserve",
		Data: ReservationConfirmedView{
			EventName: eventName,
			Name:      output.Name,
			Email:     output.Email,
			Quantity:  output.Quantity,
			Token:     output.Token,
			Message:   output.Message,
			Status:    output.Status,
		},
	})
}

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
	defer func() { _ = tx.Rollback() }()

	var eventID int
	var status string

	err = tx.QueryRow(`
		SELECT event_id, status
		FROM reservations
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(&eventID, &status)

	if err != nil {
		http.Error(w, "reserva não encontrada", http.StatusNotFound)
		return
	}

	// idempotência
	if entity.ReservationStatus(status) == entity.StatusCanceled {
		_ = tx.Commit()

		h.renderTemplate(w, "layout", RenderTemplateData{
			Page: "reservation_cancel",
			Title: buildTitle("Reserva cancelada", ""),
			Data: ReservationCancelView{
				Message: "Essa reserva já foi cancelada.",
				EventID: eventID,
			},
		})
		return
	}

	_, err = tx.Exec(`
		UPDATE reservations
		SET status = $1
		WHERE token = $2
	`, string(entity.StatusCanceled), token)

	if err != nil {
		http.Error(w, "erro ao cancelar", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page: "reservation_cancel",
		Title: buildTitle("Reserva cancelada", ""),
		Data: ReservationCancelView{
			Message: "Reserva cancelada com sucesso.",
			EventID: eventID,
		},
	})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data RenderTemplateData) {
	t := template.Must(template.ParseGlob("templates/*.html"))

	if err := t.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildTitle(page string, context string) string {
	const appName = "SofReserve"

	if context != "" {
		return fmt.Sprintf("%s • %s • %s", context, page, appName)
	}

	return fmt.Sprintf("%s • %s", page, appName)
}