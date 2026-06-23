package http

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"sof-reserve/internal/core/dto"
	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
	"sof-reserve/internal/core/port"
	"sof-reserve/internal/core/usecase"
	"sof-reserve/internal/shared/id"
	"sof-reserve/internal/shared/message"
	"sof-reserve/internal/shared/security"
)

var tmpl = template.Must(
	template.ParseGlob("internal/view/templates/*.html"),
)

// =====================
// VIEW MODELS
// =====================

// =====================
// EVENT (DOMÍNIO / VIEW BASE)
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
	PublicLink    string

	OrganizerEventCount int
	OrganizerEmail      string

	WhatsAppShareText string
}

// =====================
// EVENT DASHBOARD (PRESENTATION DATA)
// =====================

type EventDashboardData struct {
	Event EventView

	PublicLink  string
	LastUpdated string

	Reservations []entity.Reservation

	WhatsAppShareText string
}

// =====================
// UI STATE (INTERAÇÃO / FEEDBACK)
// =====================

type EventDashboardUIState struct {
	ActiveTab string

	CheckinMessage string
	CheckinError   string
}

// =====================
// VIEW FINAL (COMPOSIÇÃO)
// =====================

type EventOwnerDashboardView struct {
	Title string

	Data EventDashboardData

	UI EventDashboardUIState

	LastCheckins []LastCheckin
}

// =====================
// LAST CHECKIN
// =====================

type LastCheckin struct {
	GuestName string
}

// =====================
// RESERVATION PAGE
// =====================

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

// =====================
// TEMPLATE WRAPPER
// =====================

type RenderTemplateData struct {
	Page  string
	Title string
	Data  any
}

// =====================
// RESERVATION VIEWS
// =====================

type ReservationErrorView struct {
	Message string
	Status  string
}

type ReservationCancelView struct {
	Message    string
	EventID    int
	EventName  string
	PublicLink string
}

type ReservationTicketView struct {
	Token        string
	QRCodeURL    string
	Number       int
	WhatsAppLink string
}

type ReservationConfirmedView struct {
	EventName  string
	Name       string
	Email      string
	Quantity   int
	Message    string
	Status     string
	CancelLink string
	Tickets    []ReservationTicketView
}

// =====================
// EVENT CREATE VIEW
// =====================

type EventCreateView struct {
	Name           string
	OrganizerEmail string
	TotalSeats     int
	EndsAt         string
	Error          string
}

// =====================
// TICKET VIEW
// =====================

type TicketViewData struct {
	EventName    string
	Token        string
	TicketNumber int

	TicketURL    string
	CheckinURL   string
	WhatsAppLink string

	IsCheckedIn bool
	CheckedInAt *time.Time
}

// =====================
// HANDLER
// =====================

type Handler struct {
	reserveUC       *usecase.CreateReservationUseCase
	confirmUC       *usecase.ConfirmReservationUseCase
	eventViewUC     *usecase.GetEventViewUseCase
	createEventUC   *usecase.CreateEventUseCase

	eventRepo       port.EventRepository
	reservationRepo port.ReservationRepository
	organizerStatsUC  *usecase.GetOrganizerStatsUseCase

	db *sql.DB

	checkinTicketUC *usecase.CheckinTicket
	ticketRepo    port.TicketRepository
}

// =====================
// TEMPLATE
// =====================

func (h *Handler) renderTemplate(
	w http.ResponseWriter,
	name string,
	data RenderTemplateData,
) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Println("template error:", err)
		return
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

	var errors []string

	if name == "" {
		errors = append(errors, "Nome é obrigatório")
	}

	if email == "" {
		errors = append(errors, "Email é obrigatório")
	}

	if email != "" && !strings.Contains(email, "@") {
		errors = append(errors, "Email inválido")
	}

	if err != nil {
		errors = append(errors, "Total de vagas inválido")
	}

	if totalSeats <= 0 {
		errors = append(errors, "Total de vagas deve ser maior que zero")
	}

	if len(errors) > 0 {
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

	stats, err := h.organizerStatsUC.Execute(ucView.OrganizerEmail)
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
		WhatsAppShareText: url.QueryEscape(
			"Acesse o evento: " + baseURL + "/e/" + ucView.PublicID,
		),

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
	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "reservations"
	}

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

	baseURL := "http://localhost:8080"

	eventViewMapped := mapEventView(event)

	data := EventOwnerDashboardView{
		Title: "Event Dashboard",

		Data: EventDashboardData{
			Event: eventViewMapped,

			PublicLink: fmt.Sprintf("%s/e/%s", baseURL, event.PublicID),

			LastUpdated: time.Now().Format("15:04:05"),

			Reservations: reservations,
		},

		UI: EventDashboardUIState{
			ActiveTab: tab,
		},
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_owner_dashboard",
		Title: buildTitle("Reservas", event.Name),
		Data:  data,
	})
}

func mapEventView(uc usecase.EventView) EventView {
	return EventView{
		ID:         uc.ID,
		Name:       uc.Name,
		TotalSeats: uc.TotalSeats,
		Reserved:   uc.Reserved,

		Available:  uc.Available,
		Percentage: uc.Percentage,

		RemainingText: uc.RemainingText,
		ShowAlert:     uc.ShowAlert,
		IsClosed:      uc.IsClosed,

		PublicID:    uc.PublicID,
	}
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


	stats, err := h.organizerStatsUC.Execute(ucView.OrganizerEmail)
	if err != nil {
		http.Error(w, "error loading stats", 500)
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

		OrganizerEventCount: stats.EventCount,
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

	cancelLink := baseURL + "/cancel?token=" + output.ReservationToken

	rows, err := h.db.Query(`
		SELECT
			ticket_number,
			token
		FROM reservation_tickets
		WHERE reservation_id = $1
		ORDER BY ticket_number
	`, output.ReservationID)

	if err != nil {
		http.Error(w, "erro ao carregar tickets", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var tickets []ReservationTicketView

	for rows.Next() {
		var ticket ReservationTicketView

		err := rows.Scan(
			&ticket.Number,
			&ticket.Token,
		)

		if err != nil {
			http.Error(w, "erro ao processar tickets", http.StatusInternalServerError)
			return
		}

		ticket.QRCodeURL =
			baseURL + "/ticket/" + ticket.Token

		ticketView := entity.TicketView{
			EventName:    eventName,
			Token:        ticket.Token,
			TicketNumber: ticket.Number,
		}

		ticket.WhatsAppLink =
			message.BuildTicketWhatsAppMessage(
				baseURL,
				ticketView,
			)

		tickets = append(tickets, ticket)
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "reservation_confirmed",
		Title: buildTitle("Confirmado", ""),
		Data: ReservationConfirmedView{
			EventName:  eventName,
			Name:       output.Name,
			Email:      output.Email,
			Quantity:   output.Quantity,
			Message:    output.Message,
			Status:     output.Status,
			CancelLink: cancelLink,
			Tickets:    tickets,
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

func (h *Handler) OwnerDashboard(
	w http.ResponseWriter,
	r *http.Request,
) {

	parts := strings.Split(
		strings.Trim(r.URL.Path, "/"),
		"/",
	)

	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	publicID := parts[1]

	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "reservations"
	}

	view, err := h.loadOwnerDashboard(publicID, tab)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_owner_dashboard",
		Title: buildTitle("Reservas", view.Data.Event.Name),
		Data:  view,
	})
}

func (h *Handler) TicketView(
	w http.ResponseWriter,
	r *http.Request,
) {

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	token := parts[1]

	ticket, err := h.ticketRepo.FindTicketViewByToken(token)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	baseURL := getBaseURL(r)

	view := TicketViewData{
		EventName:    ticket.EventName,
		Token:        ticket.Token,
		TicketNumber: ticket.TicketNumber,

		TicketURL:  baseURL + "/ticket/" + ticket.Token,
		CheckinURL: baseURL + "/manage/checkin?token=" + ticket.Token,
		WhatsAppLink: message.BuildTicketWhatsAppMessage(
			baseURL,
			ticket,
		),

		IsCheckedIn: ticket.CheckedInAt != nil,
		CheckedInAt: ticket.CheckedInAt,
	}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "ticket_view",
		Title: buildTitle("Ticket", ""),
		Data:  view,
	})
}

func (h *Handler) OwnerCheckin(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	parts := strings.Split(
		strings.Trim(r.URL.Path, "/"),
		"/",
	)

	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	publicID := parts[1]

	token := strings.TrimSpace(
		r.FormValue("token"),
	)

	log.Println("TOKEN RECEBIDO:", token)

	if token == "" {
		http.Error(
			w,
			"token is required",
			http.StatusBadRequest,
		)
		return
	}

	log.Println("Executando checkin para:", token)

	err := h.checkinTicketUC.Execute(token)

	log.Printf("Resultado Execute(): %v\n", err)

	view, loadErr := h.loadOwnerDashboard(
		publicID,
		"checkin",
	)

	if loadErr != nil {
		http.NotFound(w, r)
		return
	}

	view.UI.CheckinError = ""
	view.UI.CheckinMessage = ""

		switch err {

		case nil:
			view.UI.CheckinMessage = "Check-in realizado com sucesso"

			view.LastCheckins = append(view.LastCheckins, LastCheckin{
				GuestName: token,
			})

		case coreErr.ErrTicketNotFound:
			view.UI.CheckinError = "Ticket não encontrado"

		case coreErr.ErrTicketAlreadyCheckedIn:
			view.UI.CheckinError = "Ticket já utilizado"

		default:
			http.Error(
				w,
				"internal server error",
				http.StatusInternalServerError,
			)
			return
		}

	h.renderTemplate(w, "layout", RenderTemplateData{
		Page:  "event_owner_dashboard",
		Title: buildTitle("Check-in", view.Data.Event.Name),
		Data:  view,
	})
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

func (h *Handler) parseReservationInput(
	r *http.Request,
) (int, int, string, string, error) {

	eventID, err := strconv.Atoi(r.FormValue("event_id"))
	if err != nil || eventID <= 0 {
		return 0, 0, "", "", errors.New("evento inválido")
	}

	qty, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		return 0, 0, "", "", errors.New("quantidade inválida")
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))

	var validationErrors []string

	if name == "" {
		validationErrors = append(
			validationErrors,
			"Nome é obrigatório",
		)
	}

	if email == "" {
		validationErrors = append(
			validationErrors,
			"Email é obrigatório",
		)
	}

	if email != "" && !strings.Contains(email, "@") {
		validationErrors = append(
			validationErrors,
			"Email inválido",
		)
	}

	if qty <= 0 {
		validationErrors = append(
			validationErrors,
			"Quantidade deve ser maior que zero",
		)
	}

	if len(validationErrors) > 0 {
		return eventID,
			qty,
			name,
			email,
			errors.New(strings.Join(validationErrors, ", "))
	}

	return eventID, qty, name, email, nil
}

func (h *Handler) loadOwnerDashboard(
	publicID string,
	tab string,
) (EventOwnerDashboardView, error) {

	event, err := h.eventViewUC.ExecuteByPublicID(publicID)
	if err != nil {
		return EventOwnerDashboardView{}, err
	}

	reservations, err := h.reservationRepo.FindConfirmedByEventID(event.ID)
	if err != nil {
		return EventOwnerDashboardView{}, err
	}

	totalConfirmed, err := h.reservationRepo.SumByEventID(event.ID)
	if err != nil {
		return EventOwnerDashboardView{}, err
	}

	return h.buildOwnerDashboard(
		event,
		reservations,
		totalConfirmed,
		tab,
	), nil
}

func (h *Handler) buildOwnerDashboard(
	event usecase.EventView,
	reservations []entity.Reservation,
	totalConfirmed int,
	tab string,
) EventOwnerDashboardView {

	available := event.TotalSeats - totalConfirmed

	percentage := 0
	if event.TotalSeats > 0 {
		percentage = (totalConfirmed * 100) / event.TotalSeats
	}

	remainingText := ""
	if available > 0 {
		remainingText = fmt.Sprintf("Restam %d vagas", available)
	} else {
		remainingText = "Sem vagas disponíveis"
	}

	publicLink := fmt.Sprintf("http://localhost:8080/e/%s", event.PublicID)

	return EventOwnerDashboardView{
		Title: "Event Dashboard",

		Data: EventDashboardData{
			Event: EventView{
				ID:            event.ID,
				Name:          event.Name,
				TotalSeats:    event.TotalSeats,
				Reserved:      totalConfirmed,
				Available:     available,
				Percentage:    percentage,
				RemainingText: remainingText,
				IsClosed:      event.IsClosed,
				PublicID:      event.PublicID,
				PublicLink:    publicLink,
				LastUpdated:   time.Now().Format("15:04:05"),
			},

			Reservations: reservations,

			WhatsAppShareText: url.QueryEscape(
				"Acesse o evento: " + publicLink,
			),

		},

		UI: EventDashboardUIState{
			ActiveTab: tab,
		},
	}
}