package handlers

import "net/http"

func (h *Handler) PaymentsSummaryHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payments summary retrieved successfully"))
}
