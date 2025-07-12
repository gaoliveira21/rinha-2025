package clients

import "os"

type PaymentProcessorDefault struct {
	baseUrl string
}

func NewPaymentProcessorDefault() *PaymentProcessorDefault {
	baseUrl := os.Getenv("PAYMENT_PROCESSOR_DEFAULT_URL")
	return &PaymentProcessorDefault{
		baseUrl: baseUrl,
	}
}

func (p *PaymentProcessorDefault) Health() *HealthCheckOutput {
	return healthCheck(p.baseUrl)
}

func (p *PaymentProcessorDefault) ProcessPayment(input *PaymentInput) error {
	return processPayment(p.baseUrl, input)
}
