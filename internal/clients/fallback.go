package clients

import (
	"os"
)

func NewPaymentProcessorFallback() *PaymentProcessor {
	baseUrl := os.Getenv("PAYMENT_PROCESSOR_FALLBACK_URL")
	return &PaymentProcessor{
		baseUrl: baseUrl,
		health:  true,
	}
}
