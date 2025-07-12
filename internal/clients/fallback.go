package clients

import "os"

type PaymentProcessorFallback struct {
	baseUrl string
}

func NewPaymentProcessorFallback() *PaymentProcessorFallback {
	baseUrl := os.Getenv("PAYMENT_PROCESSOR_FALLBACK_URL")
	return &PaymentProcessorFallback{
		baseUrl: baseUrl,
	}
}

func (p *PaymentProcessorFallback) Health() *HealthCheckOutput {
	return healthCheck(p.baseUrl)
}

func (p *PaymentProcessorFallback) ProcessPayment(input *PaymentInput) error {
	return processPayment(p.baseUrl, input)
}
