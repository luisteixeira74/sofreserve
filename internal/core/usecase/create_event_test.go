package usecase_test

import (
	"testing"
	"time"

	"sof-reserve/internal/core/entity"
	"sof-reserve/internal/core/usecase"
)

/*
=========================================
TESTE: CREATE EVENT USECASE
=========================================

O que estamos validando:
1. Evento é criado corretamente
2. Nome é persistido
3. Total de vagas é persistido
4. Data de evento é aceita
5. ID é gerado

Esse é o core do sistema SOF_RESERVE.
=========================================
*/

type InMemoryEventRepository struct {
	events map[int]entity.Event
	nextID int
}

func NewInMemoryEventRepository() *InMemoryEventRepository {
	return &InMemoryEventRepository{
		events: make(map[int]entity.Event),
		nextID: 1,
	}
}

func (r *InMemoryEventRepository) Create(name string, totalSeats int, endsAt time.Time) (int, error) {
	id := r.nextID
	r.nextID++

	r.events[id] = entity.Event{
		ID:         id,
		Name:       name,
		TotalSeats: totalSeats,
		EndsAt:     endsAt,
	}

	return id, nil
}

func TestCreateEvent(t *testing.T) {
	// ARRANGE
	repo := NewInMemoryEventRepository()
	uc := usecase.NewCreateEventUseCase(repo)

	eventName := "Show Rock"
	totalSeats := 100
	eventDate := time.Now().Add(24 * time.Hour)

	// ACT
	id, err := uc.Execute(eventName, totalSeats, eventDate)

	// ASSERT
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id == 0 {
		t.Error("expected valid event ID")
	}

	// valida no repo
	event, exists := repo.events[id]

	if !exists {
		t.Fatal("event not found in repository")
	}

	if event.Name != eventName {
		t.Errorf("expected %s, got %s", eventName, event.Name)
	}

	if event.TotalSeats != totalSeats {
		t.Errorf("expected %d seats, got %d", totalSeats, event.TotalSeats)
	}
}