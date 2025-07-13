package handlers

import (
	"rinha2025/internal/clients"
	"rinha2025/internal/queue"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	dispatcher               *queue.Dispatcher
	pool                     *pgxpool.Pool
	paymentProcessorDefault  *clients.PaymentProcessor
	paymentProcessorFallback *clients.PaymentProcessor
}

type ErrorResp struct {
	Message string `json:"message"`
}

func NewHandler(
	pool *pgxpool.Pool,
	d *queue.Dispatcher,
	paymentProcessorDefault *clients.PaymentProcessor,
	paymentProcessorFallback *clients.PaymentProcessor,
) *Handler {
	return &Handler{
		dispatcher:               d,
		pool:                     pool,
		paymentProcessorDefault:  paymentProcessorDefault,
		paymentProcessorFallback: paymentProcessorFallback,
	}
}
