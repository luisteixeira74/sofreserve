package main

import (
	"log"
	"net/http"

	httpadapter "sof-reserve/internal/adapter/http"
	"sof-reserve/internal/adapter/repository/postgres"
	"sof-reserve/internal/core/usecase"
	"sof-reserve/internal/infra/db"
)

func main() {
	// =====================
	// INFRA (DB)
	// =====================
	database := db.NewConnection()

	// =====================
	// REPOSITORIES
	// =====================
	eventRepo := postgres.NewEventRepository(database)
	reservationRepo := postgres.NewReservationRepository(database)

	// =====================
	// USECASES
	// =====================
	reserveUseCase := usecase.NewReserveSpotUseCase(
		eventRepo,
		reservationRepo,
	)

	confirmationUseCase := usecase.NewConfirmReservationUseCase(
		database,
	)

	// =====================
	// ROUTER
	// =====================
	router := httpadapter.NewRouter(reserveUseCase, confirmationUseCase, database)

	// =====================
	// SERVER
	// =====================
	log.Println("Server running on :8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}