package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HealthCheckOutput struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

func healthCheck(baseUrl string) *HealthCheckOutput {
	out := &HealthCheckOutput{
		Failing:         true,
		MinResponseTime: 0,
	}
	resp, err := http.Get(baseUrl + "/payments/service-health")
	if err != nil {
		return out
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		log.Printf("Error decoding health check response from %s: %v", baseUrl, err)
		return out
	}

	return out
}

type PaymentInput struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

func processPayment(baseUrl string, input *PaymentInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(body)
	resp, err := http.Post(baseUrl+"/payments/service-health", "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, baseUrl)
	}

	return nil
}
