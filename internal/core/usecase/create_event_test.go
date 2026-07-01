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
Valida criação de evento com:
- nome
- total de vagas
- data final
- persistência no repositório
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
		ID:         int64(id),
		Name:       name,
		TotalSeats: totalSeats,
		EndsAt:     endsAt,
	}

	return id, nil
}

func TestCreateEventUseCase(t *testing.T) {
	// ARRANGE
	repo := NewInMemoryEventRepository()
	uc := usecase.NewCreateEventUseCase(repo)

	eventName := "Show Rock"
	totalSeats := 100
	eventDate := time.Now().Add(24 * time.Hour)

	input := usecase.CreateEventInput{
		Name:       eventName,
		TotalSeats: totalSeats,
		EndsAt:     eventDate,
	}

	// ACT
	id, err := uc.Execute(input)

	// ASSERT
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id <= 0 {
		t.Fatal("expected valid event ID")
	}

	event, exists := repo.events[id]
	if !exists {
		t.Fatal("event not found in repository")
	}

	if event.Name != eventName {
		t.Errorf("expected name %s, got %s", eventName, event.Name)
	}

	if event.TotalSeats != totalSeats {
		t.Errorf("expected %d seats, got %d", totalSeats, event.TotalSeats)
	}
}