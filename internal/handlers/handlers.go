package handlers

import (
	"rinha2025/internal/clients"
	"rinha2025/internal/dao"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	paymentsDAO              *dao.PaymentsDAO
	paymentProcessorDefault  *clients.PaymentProcessor
	paymentProcessorFallback *clients.PaymentProcessor
}

type ErrorResp struct {
	Message string `json:"message"`
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		paymentsDAO:              dao.NewPaymentsDAO(pool),
		paymentProcessorDefault:  clients.NewPaymentProcessorDefault(),
		paymentProcessorFallback: clients.NewPaymentProcessorFallback(),
	}
}

func (h *Handler) GetPaymentProcessorDefault() *clients.PaymentProcessor {
	return h.paymentProcessorDefault
}

func (h *Handler) GetPaymentProcessorFallback() *clients.PaymentProcessor {
	return h.paymentProcessorFallback
}
