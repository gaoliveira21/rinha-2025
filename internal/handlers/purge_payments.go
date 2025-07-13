package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) PurgePaymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	h.dispatcher.Clear()
	conn, err := h.pool.Acquire(context.Background())
	if err != nil {
		log.Printf("Failed to acquire connection: %v\n", err)
		resp := ErrorResp{Message: "Failed to purge payments"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), `TRUNCATE payments_default, payments_fallback`)
	if err != nil {
		log.Printf("Failed to truncate payments tables: %v\n", err)
		resp := ErrorResp{Message: "Failed to purge payments"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payment request accepted"})
}
