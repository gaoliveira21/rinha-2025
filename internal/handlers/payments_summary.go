package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (h *Handler) PaymentsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	from := r.URL.Query().Get("from")
	if from == "" {
		from = "1970-01-01T00:00:00Z"
	}
	to := r.URL.Query().Get("to")
	if to == "" {
		to = time.Now().Format(time.RFC3339)
	}

	summary, err := h.paymentsDAO.GetPaymentsSummary(from, to)
	if err != nil {
		log.Printf("Error getting default payment summary: %v\n", err)
		resp := ErrorResp{Message: "Failed to get default payment summary"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}
