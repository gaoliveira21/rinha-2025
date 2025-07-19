package queue

import (
	"context"
	"log"
	"rinha2025/internal/clients"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentWorker struct {
	ctx                      context.Context
	pool                     *pgxpool.Pool
	dispatcher               *Dispatcher
	paymentProcessorDefault  *clients.PaymentProcessor
	paymentProcessorFallback *clients.PaymentProcessor
}

func NewPaymentWorker(
	pool *pgxpool.Pool,
	paymentProcessorDefault *clients.PaymentProcessor,
	paymentProcessorFallback *clients.PaymentProcessor,
	ctx context.Context,
) *PaymentWorker {
	return &PaymentWorker{
		ctx:                      ctx,
		pool:                     pool,
		paymentProcessorDefault:  paymentProcessorDefault,
		paymentProcessorFallback: paymentProcessorFallback,
	}
}

func (w *PaymentWorker) SetDispatcher(d *Dispatcher) {
	w.dispatcher = d
}

func (w *PaymentWorker) ProcessPayment(job *PaymentJob) {
	job.Attempts++

	if job.Attempts > 3 {
		conn, err := w.pool.Acquire(w.ctx)
		if err != nil {
			log.Printf("Failed to acquire connection: %v", err)
			return
		}
		defer conn.Release()

		_, err = conn.Exec(w.ctx, `INSERT INTO failed_payments_queue (correlation_id, amount) VALUES ($1, $2)`,
			job.CorrelationID, job.Amount)
		if err != nil {
			log.Printf("Failed to insert payment into db queue: %v", err)
		}
		return
	}

	if w.paymentProcessorDefault.IsHealthy() {
		requestedAt := time.Now()
		err := w.paymentProcessorDefault.ProcessPayment(&clients.PaymentInput{
			CorrelationID: job.CorrelationID,
			Amount:        job.Amount,
			RequestedAt:   requestedAt,
		})
		if err != nil {
			w.dispatcher.Enqueue(job)
			return
		}

		conn, err := w.pool.Acquire(w.ctx)
		if err != nil {
			log.Printf("Failed to acquire connection: %v", err)
			return
		}
		defer conn.Release()

		_, err = conn.Exec(w.ctx, `INSERT INTO payments_default (correlation_id, amount, requested_at) VALUES ($1, $2, $3)`,
			job.CorrelationID, job.Amount, requestedAt)
		if err != nil {
			log.Printf("Failed to insert payment default: %v", err)
		}
		return
	}

	w.dispatcher.Enqueue(job)
}
