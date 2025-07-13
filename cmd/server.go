package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rinha2025/internal/clients"
	"rinha2025/internal/handlers"
	"rinha2025/internal/queue"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	port := os.Getenv("PORT")

	server := &http.Server{
		Addr: ":" + port,
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	paymentPcrDft := clients.NewPaymentProcessorDefault()
	paymentPcrFbk := clients.NewPaymentProcessorFallback()

	go paymentPcrDft.HearthBeat()
	go paymentPcrFbk.HearthBeat()

	worker := queue.NewPaymentWorker(pool, paymentPcrDft, paymentPcrFbk, ctx)
	d := queue.NewDispatcher(
		worker,
		20,
		2048,
	)
	worker.SetDispatcher(d)

	go d.Start(ctx)

	handler := handlers.NewHandler(pool, d, paymentPcrDft, paymentPcrFbk)
	http.HandleFunc("POST /payments", handler.PaymentsHandler)
	http.HandleFunc("POST /purge-payments", handler.PurgePaymentsHandler)
	http.HandleFunc("GET /payments-summary", handler.PaymentsSummaryHandler)
	http.HandleFunc("GET /health", handler.HealthCheckHandler)

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

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped")
}
