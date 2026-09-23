package usecase_test

import (
	"testing"
	"time"

	"sof-reserve/internal/core/entity"
	coreErr "sof-reserve/internal/core/errors"
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
	events map[int64]entity.Event
	nextID int64
}

func NewInMemoryEventRepository() *InMemoryEventRepository {
	return &InMemoryEventRepository{
		events: make(map[int64]entity.Event),
		nextID: 1,
	}
}

func (r *InMemoryEventRepository) Create(event entity.Event) (int64, error) {
	id := r.nextID
	r.nextID++

	event.ID = int64(id)
	r.events[event.ID] = event

	return event.ID, nil
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

	event, exists := repo.FindByID(id)
	if !exists {
		t.Fatal("event not found in repository")
	}

	if event.ID <= 0 {
		t.Fatal("expected valid event ID")
	}

	if event.Name != eventName {
		t.Errorf("expected name %s, got %s", eventName, event.Name)
	}

	if event.TotalSeats != totalSeats {
		t.Errorf("expected %d seats, got %d", totalSeats, event.TotalSeats)
	}
}

func TestCreateEventUseCase_InvalidName(t *testing.T) {
	// ARRANGE
	repo := NewInMemoryEventRepository()
	uc := usecase.NewCreateEventUseCase(repo)

	input := usecase.CreateEventInput{
		Name:       "",
		TotalSeats: 100,
		EndsAt:     time.Now().Add(24 * time.Hour),
	}

	// ACT
	_, err := uc.Execute(input)

	// ASSERT
	if err != coreErr.ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestCreateEventUseCase_InvalidQuantity(t *testing.T) {
	// ARRANGE
	repo := NewInMemoryEventRepository()
	uc := usecase.NewCreateEventUseCase(repo)

	input := usecase.CreateEventInput{
		Name:       "Show Rock",
		TotalSeats: 0,
		EndsAt:     time.Now().Add(24 * time.Hour),
	}

	// ACT
	_, err := uc.Execute(input)

	// ASSERT
	if err != coreErr.ErrInvalidQuantity {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}
}

func TestCreateEventUseCase_EventClosed(t *testing.T) {
	// ARRANGE
	repo := NewInMemoryEventRepository()
	uc := usecase.NewCreateEventUseCase(repo)

	input := usecase.CreateEventInput{
		Name:       "Show Rock",
		TotalSeats: 100,
		EndsAt:     time.Now().Add(-24 * time.Hour),
	}

	// ACT
	_, err := uc.Execute(input)

	// ASSERT
	if err != coreErr.ErrEventClosed {
		t.Fatalf("expected ErrEventClosed, got %v", err)
	}
}

func (r *InMemoryEventRepository) FindByID(id int64) (entity.Event, bool) {
	event, ok := r.events[id]
	return event, ok
}
