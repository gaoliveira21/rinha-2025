package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rinha2025/internal/handlers"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	port := os.Getenv("PORT")

	server := &http.Server{
		Addr: ":" + port,
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	handler := handlers.NewHandler(pool)
	http.HandleFunc("POST /payments", handler.PaymentsHandler)
	http.HandleFunc("GET /payments-summary", handler.PaymentsSummaryHandler)

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving on port", port)
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)
	<-sc

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	pool.Close()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped")
}
