package http

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"

	"sof-reserve/internal/core/dto"
	"sof-reserve/internal/core/usecase"
)

type EventView struct {
	ID         int
	Name       string
	TotalSeats int
	Reserved   int
	Available  int
}

type Handler struct {
	db        *sql.DB
	reserveUC *usecase.ReserveSpotUseCase
}

// =====================
// HEALTH
// =====================

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// =====================
// CREATE EVENT
// =====================

func (h *Handler) CreateEventPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/create_event.html"))
	_ = tmpl.Execute(w, nil)
}

func (h *Handler) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	totalSeatsStr := r.FormValue("total_seats")
	totalSeats, err := strconv.Atoi(totalSeatsStr)
	if err != nil || totalSeats <= 0 {
		http.Error(w, "invalid total_seats", http.StatusBadRequest)
		return
	}

	var id int
	err = h.db.QueryRow(
		"INSERT INTO events (name, total_seats) VALUES ($1, $2) RETURNING id",
		name, totalSeats,
	).Scan(&id)

	if err != nil {
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

	err = h.db.QueryRow(
		"SELECT name, total_seats FROM events WHERE id = $1",
		id,
	).Scan(&name, &totalSeats)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	var reserved int
	err = h.db.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM reservations WHERE event_id = $1",
		id,
	).Scan(&reserved)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := EventView{
		ID:         id,
		Name:       name,
		TotalSeats: totalSeats,
		Reserved:   reserved,
		Available:  totalSeats - reserved,
	}

	tmpl := template.Must(template.ParseFiles("templates/event.html"))
	_ = tmpl.Execute(w, view)
}

// =====================
// CREATE RESERVATION
// =====================

func (h *Handler) CreateReservationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventID, err := strconv.Atoi(r.FormValue("event_id"))
	if err != nil || eventID <= 0 {
		http.Error(w, "invalid event_id", http.StatusBadRequest)
		return
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || qty <= 0 {
		http.Error(w, "invalid quantity", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	req := dto.ReserveRequest{
		EventID:  eventID,
		Name:     name,
		Quantity: qty,
	}

	if err := h.reserveUC.Execute(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/evento?id="+strconv.Itoa(eventID), http.StatusSeeOther)
}