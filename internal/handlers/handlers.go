package handlers

import "rinha2025/internal/clients"

type Handler struct {
	paymentProcessorDefault  *clients.PaymentProcessorDefault
	paymentProcessorFallback *clients.PaymentProcessorFallback
}

func NewHandler() *Handler {
	return &Handler{
		paymentProcessorDefault:  clients.NewPaymentProcessorDefault(),
		paymentProcessorFallback: clients.NewPaymentProcessorFallback(),
	}
}
