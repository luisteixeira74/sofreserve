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
	ID         int
	Name       string
	TotalSeats int
	Reserved   int
	Available  int
	Percentage int
	ColorClass string
	RemainingText string
	ShowAlert     bool
	IsClosed      bool
}

type Handler struct {
	db        *sql.DB
	reserveUC *usecase.ReserveSpotUseCase
	confirmUC *usecase.ConfirmReservationUseCase
}

type ReservationPageData struct {
	Error    string
	EventID  int
	Name     string
	Email    string
	Quantity int
}

// =====================
// HELPERS
// =====================

// func (h *Handler) renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
// 	t, err := template.ParseFiles("templates/" + tmpl)
// 	if err != nil {
// 		http.Error(w, "erro ao carregar template", http.StatusInternalServerError)
// 		return
// 	}

// 	_ = t.Execute(w, data)
// }

// =====================
// HEALTH
// =====================

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// =====================
// CREATE EVENT
// =====================

func (h *Handler) CreateEventPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "layout", map[string]any{
		"Page": "create_event",
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
	if name == "" || err != nil || totalSeats <= 0 {
		h.renderTemplate(w, "create_event.html", map[string]interface{}{
			"Error": "Dados inválidos",
			"Name":  name,
		})
		return
	}

	// validação extra para evitar eventos com muitas vagas
	if totalSeats <= 0 || totalSeats > 1000 {
		h.renderTemplate(w, "create_event.html", map[string]interface{}{
			"Error": "O evento pode ter no máximo 1000 vagas",
		})
		return
	}

	endsAtStr := r.FormValue("ends_at")

	endsAt, err := time.Parse("2006-01-02", endsAtStr)
	if err != nil {
		h.renderTemplate(w, "create_event.html", map[string]interface{}{
			"Error": "Data inválida",
		})
		return
	}

	var id int
	err = h.db.QueryRow(
		"INSERT INTO events (name, total_seats, ends_at) VALUES ($1, $2, $3) RETURNING id",
		name, totalSeats, endsAt,
	).Scan(&id)

	if err != nil {
		// http.Error(w, "erro ao criar evento", http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/evento?id="+strconv.Itoa(id), http.StatusSeeOther)
}

// =====================
// EVENT PAGE
// =====================

func (h *Handler) EventPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	var name string
	var totalSeats int
	var endsAt time.Time

	err = h.db.QueryRow(
		"SELECT name, total_seats, ends_at FROM events WHERE id = $1",
		id,
	).Scan(&name, &totalSeats, &endsAt)

	if err != nil {
		http.NotFound(w, r)
		return
	}

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
	err = h.db.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM reservations WHERE event_id = $1 AND status = 'confirmed'",
		id,
	).Scan(&reserved)

	percentage := 0
	if totalSeats > 0 {
		percentage = (reserved * 100) / totalSeats
	}

	if err != nil {
		http.Error(w, "erro ao carregar reservas", http.StatusInternalServerError)
		return
	}

	colorClass := "bg-green-500"

	if percentage < 50 {
		colorClass = "bg-green-500"
	} else if percentage < 80 {
		colorClass = "bg-yellow-500"
	} else {
		colorClass = "bg-red-500"
	}

	view := EventView{
		ID:         id,
		Name:       name,
		TotalSeats: totalSeats,
		Reserved:   reserved,
		Available:  totalSeats - reserved,
		Percentage: percentage,
		ColorClass: colorClass,
		RemainingText: remainingText,
		ShowAlert:     showAlert,
		IsClosed:      isClosed,
	}

	h.renderTemplate(w, "event.html", view)
}

// =====================
// CREATE RESERVATION
// =====================

func (h *Handler) CreateReservationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventIDStr := r.FormValue("event_id")
	qtyStr := r.FormValue("quantity")

	eventID, errID := strconv.Atoi(eventIDStr)
	qty, errQty := strconv.Atoi(qtyStr)

	name := r.FormValue("name")
	email := r.FormValue("email")

	data := ReservationPageData{
		EventID:  eventID,
		Name:     name,
		Email:    email,
		Quantity: qty,
	}

	// validação básica
	if errID != nil || eventID <= 0 {
		data.Error = "Evento inválido"
		h.renderTemplate(w, "reservation.html", data)
		return
	}

	if errQty != nil || qty <= 0 {
		data.Error = "Quantidade inválida"
		h.renderTemplate(w, "reservation.html", data)
		return
	}

	if qty <= 0 || qty > 10 {
		data.Error = "Você pode reservar no máximo 10 vagas"
		h.renderTemplate(w, "reservation.html", data)
		return
	}

	if name == "" {
		data.Error = "Nome é obrigatório"
		h.renderTemplate(w, "reservation.html", data)
		return
	}

	if email == "" || !strings.Contains(email, "@") {
		data.Error = "Email inválido"
		h.renderTemplate(w, "reservation.html", data)
		return
	}

	req := dto.ReserveRequest{
		EventID:  eventID,
		Name:     name,
		Email:    email,
		Quantity: qty,
	}

	err := h.reserveUC.Execute(req)
	if err != nil {

		var appErr *coreErr.AppError

		if errors.As(err, &appErr) {
			data.Error = appErr.Message

			// mantém dados preenchidos
			data.EventID = eventID
			data.Name = name
			data.Email = email
			data.Quantity = qty

			h.renderTemplate(w, "reservation.html", data)
			return
		}
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/evento?id="+strconv.Itoa(eventID), http.StatusSeeOther)
}

func (h *Handler) ReservationPage(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.URL.Query().Get("event_id")
	eventID, _ := strconv.Atoi(eventIDStr)

	data := ReservationPageData{
		EventID: eventID,
	}

	h.renderTemplate(w, "reservation.html", data)
}

func (h *Handler) ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	output, err := h.confirmUC.Execute(token)
	if err != nil {

		var appErr *coreErr.AppError
		if errors.As(err, &appErr) {
			h.renderTemplate(w, "confirm.html", map[string]interface{}{
				"Status":  "error",
				"Message": appErr.Message,
			})
			return
		}

		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, "confirm.html", output)
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

	var name, email string
	var quantity, eventID int
	var status string

	err = tx.QueryRow(`
		SELECT name, email, quantity, event_id, status
		FROM reservations
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(&name, &email, &quantity, &eventID, &status)

	if err != nil {
		http.Error(w, "reserva não encontrada", http.StatusNotFound)
		return
	}

	// idempotência
	if entity.ReservationStatus(status) == entity.StatusCanceled {
		_ = tx.Commit()

		h.renderTemplate(w, "cancel.html", map[string]interface{}{
			"Message": "Essa reserva já foi cancelada.",
			"EventID": eventID,
		})
		return
	}

	// cancela
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

	h.renderTemplate(w, "cancel.html", map[string]interface{}{
		"Message": "Reserva cancelada com sucesso.",
		"EventID": eventID,
	})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	t := template.Must(template.ParseGlob("templates/*.html"))

	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)	
	}
}