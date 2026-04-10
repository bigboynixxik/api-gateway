package handlers

import (
	"api-gateway/pkg/response"
	"net/http"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
