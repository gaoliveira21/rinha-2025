package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type PaymentProcessor struct {
	baseUrl string
	health  bool
	mu      sync.Mutex
}

type HealthCheckOutput struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

func (p *PaymentProcessor) HealthCheck() *HealthCheckOutput {
	out := &HealthCheckOutput{
		Failing:         true,
		MinResponseTime: 0,
	}
	resp, err := http.Get(p.baseUrl + "/payments/service-health")
	if err != nil {
		log.Printf("Error fetching health check from %s: %v\n", p.baseUrl, err)
		return out
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return out
	}

	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		log.Printf("Error decoding health check response from %s: %v", p.baseUrl, err)
		return out
	}

	return out
}

func (p *PaymentProcessor) HearthBeat(svcChan chan struct{}) {
	for {
		output := p.HealthCheck()
		p.mu.Lock()
		p.health = !output.Failing
		p.mu.Unlock()

		if p.health {
			select {
			case svcChan <- struct{}{}:
			default:
				// Channel is full, skip sending to avoid blocking
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func (p *PaymentProcessor) IsHealthy() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.health
}

type PaymentInput struct {
	CorrelationID string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}

func (p *PaymentProcessor) ProcessPayment(input *PaymentInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(body)
	resp, err := http.Post(p.baseUrl+"/payments", "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, p.baseUrl)
	}

	return nil
}
