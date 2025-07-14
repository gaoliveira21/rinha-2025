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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	svcChan := make(chan struct{}, 100)
	defer cancel()
	port := os.Getenv("PORT")

	server := &http.Server{
		Addr: ":" + port,
	}

	config, _ := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	config.MaxConns = 22
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	paymentPcrDft := clients.NewPaymentProcessorDefault()
	paymentPcrFbk := clients.NewPaymentProcessorFallback()

	go paymentPcrDft.HearthBeat(svcChan)
	go paymentPcrFbk.HearthBeat(svcChan)

	worker := queue.NewPaymentWorker(pool, paymentPcrDft, paymentPcrFbk, ctx)
	d := queue.NewDispatcher(
		worker,
		20,
		6048,
	)
	worker.SetDispatcher(d)

	go d.Start(ctx)

	go processFailedPayment(pool, svcChan, d, ctx)

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
	close(svcChan)
	log.Println("Server gracefully stopped")
}

func processFailedPayment(pool *pgxpool.Pool, svcChan chan struct{}, d *queue.Dispatcher, ctx context.Context) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to acquire connection: %v", err)
		return
	}
	defer conn.Release()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping processing of failed payments")
			return
		case _, ok := <-svcChan:
			if !ok {
				log.Println("Service channel closed, stopping processing of failed payments")
				return
			}

			log.Println("Processing failed payments")
			var jobs []*queue.PaymentJob
			var toDelete []string
			tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
			if err != nil {
				log.Printf("Failed to begin transaction: %v", err)
				continue
			}

			rows, err := tx.Query(
				ctx,
				`SELECT correlation_id, amount FROM failed_payments_queue LIMIT 400`,
			)
			if err != nil {
				tx.Rollback(ctx)
				if errors.Is(err, pgx.ErrNoRows) {
					log.Println("No failed payments to process")
					continue
				}
				log.Printf("Failed to query failed payments: %v", err)
				continue
			}

			for rows.Next() {
				var correlationID string
				var amount float64
				if err := rows.Scan(&correlationID, &amount); err != nil {
					log.Printf("Failed to scan failed payment job: %v", err)
					continue
				}

				jobs = append(jobs, &queue.PaymentJob{
					CorrelationID: correlationID,
					Amount:        amount,
					Attempts:      0,
				})
				toDelete = append(toDelete, correlationID)
			}

			rows.Close()

			_, err = tx.Exec(ctx, `DELETE FROM failed_payments_queue WHERE correlation_id = ANY ($1)`, toDelete)
			if err != nil {
				tx.Rollback(ctx)
				log.Printf("Failed to delete failed payments from db queue: %v", err)
				continue
			}

			tx.Commit(ctx)

			for _, job := range jobs {
				d.Enqueue(job)
				log.Printf("Re-enqueued failed payment job %s for processing", job.CorrelationID)
			}
		}
	}
}
