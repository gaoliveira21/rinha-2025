package handlers

import (
	"encoding/json"
	"net/http"
)

type PaymentBody struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

func (h *Handler) PaymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body PaymentBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{Message: "Invalid request body"})
		return
	}

	if h.paymentProcessorDefault.IsHealthy() {
		//TODO: Send to the default payment processor queue
	}

	if h.paymentProcessorFallback.IsHealthy() {
		// TODO: Send to the fallback payment processor queue
	}

	// TODO: Insert payment into the database queue to be processed later
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payment request accepted"})
}
