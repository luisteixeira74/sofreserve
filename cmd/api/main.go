package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "sof-reserve/internal/adapter/http"
	"sof-reserve/internal/adapter/repository/postgres"
	"sof-reserve/internal/config"
	"sof-reserve/internal/core/usecase"
	"sof-reserve/internal/infra/db"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	// =====================
	// INFRA (DB)
	// =====================
	database := db.NewConnection(cfg)
	defer database.Close()

	// =====================
	// PORT
	// =====================
	port := cfg.Port

	// =====================
	// REPOSITORIES
	// =====================
	eventRepo := postgres.NewEventRepository(database)
	reservationRepo := postgres.NewReservationRepository(database)
	ticketRepo := postgres.NewTicketRepository(database)

	// =====================
	// USECASES
	// =====================
	clock := usecase.RealClock{}

	reserveUseCase := usecase.NewCreateReservationUseCase(
		database,
		eventRepo,
		reservationRepo,
		clock,
	)

	confirmationUseCase := usecase.NewConfirmReservationUseCase(
		database,
	)

	eventViewUC := usecase.NewGetEventViewUseCase(
		eventRepo,
		reservationRepo,
		clock,
	)

	createEventUC := usecase.NewCreateEventUseCase(
		eventRepo,
	)

	getOrganizerStatsUC := usecase.NewGetOrganizerStatsUseCase(
		eventRepo,
	)

	

	checkinTicketUC := usecase.NewCheckinTicket(
		database,
		ticketRepo,
	)

	// =====================
	// ROUTER
	// =====================
	router := httpadapter.NewRouter(
		reserveUseCase,
		confirmationUseCase,
		eventViewUC,
		eventRepo,
		reservationRepo,
		ticketRepo,
		createEventUC,
		getOrganizerStatsUC,
		checkinTicketUC,
		database,

	)



	// =====================
	// SERVER CONFIG
	// =====================
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// =====================
	// START SERVER (goroutine)
	// =====================
	go func() {
		log.Println("Server running on port:", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error:", err)
		}
	}()

	// =====================
	// GRACEFUL SHUTDOWN
	// =====================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown:", err)
	}

	log.Println("Server exited properly")
}