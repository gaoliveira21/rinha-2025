package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthCheckResp struct {
	Health                         bool `json:"health"`
	PaymentProcessorDefaultHealth  bool `json:"payment_processor_default_health"`
	PaymentProcessorFallbackHealth bool `json:"payment_processor_fallback_health"`
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := &HealthCheckResp{
		Health:                         true,
		PaymentProcessorDefaultHealth:  h.paymentProcessorDefault.IsHealthy(),
		PaymentProcessorFallbackHealth: h.paymentProcessorFallback.IsHealthy(),
	}
	json.NewEncoder(w).Encode(resp)
}
