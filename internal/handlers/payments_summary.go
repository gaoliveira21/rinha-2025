package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type PaymentSummary struct {
	TotalRequests int64   `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentsSummaryResp struct {
	Default  *PaymentSummary `json:"default"`
	Fallback *PaymentSummary `json:"fallback"`
}

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

	conn, err := h.pool.Acquire(context.Background())
	if err != nil {
		log.Printf("Failed to acquire connection: %v\n", err)
		resp := ErrorResp{Message: "Failed to get payments summary"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
	}
	defer conn.Release()

	defaultSum := &PaymentSummary{}
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM payments_default
		WHERE requested_at >= $1 AND requested_at <= $2
	`, from, to).Scan(&defaultSum.TotalRequests, &defaultSum.TotalAmount)
	if err != nil {
		log.Printf("Failed to query payments default: %v\n", err)
		resp := ErrorResp{Message: "Failed to get payments summary"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
	}

	fallbackSum := &PaymentSummary{}
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM payments_fallback
		WHERE requested_at >= $1 AND requested_at <= $2
	`, from, to).Scan(&fallbackSum.TotalRequests, &fallbackSum.TotalAmount)
	if err != nil {
		log.Printf("Failed to query payments fallback: %v\n", err)
		resp := ErrorResp{Message: "Failed to get payments summary"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
	}

	out := &PaymentsSummaryResp{
		Default:  defaultSum,
		Fallback: fallbackSum,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}
