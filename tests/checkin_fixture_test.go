package tests

import (
	"sof-reserve/internal/config"
	"sof-reserve/internal/infra/db"
	"testing"
	"time"
)

const (
	checkinToken    = "tkt_TESTE_123"
	checkinPublicID = "djids3wi"
)

func setupCheckinFixture(t *testing.T) {
	t.Helper()

	cfg := config.Load()
	database := db.NewConnection(cfg)
	defer database.Close()

	// Limpa dados de teste anteriores.
	_, err := database.Exec(`
		DELETE FROM reservation_tickets
		WHERE token = $1
	`, checkinToken)
	if err != nil {
		t.Fatal(err)
	}

	_, err = database.Exec(`
		DELETE FROM reservations
		WHERE token = $1
	`, "reservation_test")
	if err != nil {
		t.Fatal(err)
	}

	// Cria o evento utilizado pelo teste.
	var eventID int64

	err = database.QueryRow(`
		INSERT INTO events (
			name,
			organizer_email,
			total_seats,
			ends_at,
			public_id,
			owner_token
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		"Race Test Event",
		"test@example.com",
		100,
		time.Now().Add(24*time.Hour),
		checkinPublicID,
		"owner_test",
	).Scan(&eventID)

	if err != nil {
		t.Fatal(err)
	}

	// Cria uma reservation para o ticket.
	var reservationID int64

	err = database.QueryRow(`
		INSERT INTO reservations (
			event_id,
			name,
			email,
			quantity,
			status,
			token,
			confirmed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id
	`,
		eventID,
		"Race Test User",
		"test@example.com",
		1,
		"confirmed",
		"reservation_test",
	).Scan(&reservationID)

	if err != nil {
		t.Fatal(err)
	}

	// Cria o ticket que será disputado pelas requisições concorrentes.
	_, err = database.Exec(`
		INSERT INTO reservation_tickets (
			reservation_id,
			event_id,
			ticket_number,
			token
		)
		VALUES ($1, $2, $3, $4)
	`,
		reservationID,
		eventID,
		1,
		checkinToken,
	)

	if err != nil {
		t.Fatal(err)
	}
}
