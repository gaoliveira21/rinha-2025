package clients

import (
	"os"
)

func NewPaymentProcessorDefault() *PaymentProcessor {
	baseUrl := os.Getenv("PAYMENT_PROCESSOR_DEFAULT_URL")
	return &PaymentProcessor{
		baseUrl: baseUrl,
	}
}
