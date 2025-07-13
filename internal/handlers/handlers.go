package handlers

import (
	"rinha2025/internal/clients"
	"rinha2025/internal/dao"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	paymentsDAO              *dao.PaymentsDAO
	paymentProcessorDefault  *clients.PaymentProcessorDefault
	paymentProcessorFallback *clients.PaymentProcessorFallback
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		paymentsDAO:              dao.NewPaymentsDAO(pool),
		paymentProcessorDefault:  clients.NewPaymentProcessorDefault(),
		paymentProcessorFallback: clients.NewPaymentProcessorFallback(),
	}
}
