package handlers

import (
	"encoding/json"
	"net/http"
	"rinha2025/internal/queue"
)

type PaymentBody struct {
	CorrelationID string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

func (h *Handler) PaymentsHandler(w http.ResponseWriter, r *http.Request) {
	var body PaymentBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return
	}

	h.dispatcher.Enqueue(&queue.PaymentJob{
		CorrelationID: body.CorrelationID,
		Amount:        body.Amount,
	})

	w.WriteHeader(http.StatusAccepted)
}
