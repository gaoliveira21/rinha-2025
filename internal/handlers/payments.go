package handlers

import "net/http"

func (h *Handler) PaymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment processed successfully"))
}
