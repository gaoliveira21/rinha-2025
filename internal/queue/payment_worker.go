package queue

import (
	"context"
	"log"
	"rinha2025/internal/clients"
	"time"

	"github.com/jackc/pgx/v5"
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
	time.Sleep(500 * time.Millisecond * time.Duration(job.Attempts))

	job.Attempts++

	if job.Attempts > 3 {
		log.Printf("Job %s exceeded max attempts, dropping job", job.CorrelationID)
		return
	}

	conn, err := w.pool.Acquire(w.ctx)
	if err != nil {
		log.Printf("Failed to acquire connection: %v", err)
		w.dispatcher.Enqueue(job)
		return
	}
	defer conn.Release()

	tx, err := conn.BeginTx(w.ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		w.dispatcher.Enqueue(job)
		return
	}

	if w.paymentProcessorDefault.IsHealthy() {
		_, err = tx.Exec(w.ctx, `INSERT INTO payments_default (correlation_id, amount, requested_at) VALUES ($1, $2, $3)`,
			job.CorrelationID, job.Amount, job.RequestedAt)
		if err != nil {
			log.Printf("Failed to insert payment default: %v", err)
			if rollbackErr := tx.Rollback(w.ctx); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
			w.dispatcher.Enqueue(job)
			return
		}

		err = w.paymentProcessorDefault.ProcessPayment(&clients.PaymentInput{
			CorrelationID: job.CorrelationID,
			Amount:        job.Amount,
		})
		if err != nil {
			log.Printf("Failed to process payment default: %v", err)
			if rollbackErr := tx.Rollback(w.ctx); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
			w.dispatcher.Enqueue(job)
			return
		}

		if err = tx.Commit(w.ctx); err != nil {
			log.Printf("Failed to commit transaction: %v", err)
			w.dispatcher.Enqueue(job)
		}
		return
	}

	if w.paymentProcessorFallback.IsHealthy() {
		_, err = tx.Exec(w.ctx, `INSERT INTO payments_fallback (correlation_id, amount, requested_at) VALUES ($1, $2, $3)`,
			job.CorrelationID, job.Amount, job.RequestedAt)
		if err != nil {
			log.Printf("Failed to insert payment fallback: %v", err)
			if rollbackErr := tx.Rollback(w.ctx); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
			w.dispatcher.Enqueue(job)
			return
		}

		err = w.paymentProcessorFallback.ProcessPayment(&clients.PaymentInput{
			CorrelationID: job.CorrelationID,
			Amount:        job.Amount,
		})
		if err != nil {
			log.Printf("Failed to process payment fallback: %v", err)
			if rollbackErr := tx.Rollback(w.ctx); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
			w.dispatcher.Enqueue(job)
			return
		}

		if err = tx.Commit(w.ctx); err != nil {
			log.Printf("Failed to commit transaction: %v", err)
			w.dispatcher.Enqueue(job)
		}
		return
	}

	w.dispatcher.Enqueue(job)
	log.Printf("Both payment processors are unhealthy, job re-enqueued: %s", job.CorrelationID)
}
