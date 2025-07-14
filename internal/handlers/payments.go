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
	w.Header().Set("Content-Type", "application/json")

	var body PaymentBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResp{Message: "Invalid request body"})
		return
	}

	h.dispatcher.Enqueue(&queue.PaymentJob{
		CorrelationID: body.CorrelationID,
		Amount:        body.Amount,
	})

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payment request accepted"})
}
