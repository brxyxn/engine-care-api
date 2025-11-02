package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Success writes an success message as a JSON response.
func Success[T any](w http.ResponseWriter, code int, data T) {
	successJSON[T](w, code, data)
}

// successJSON writes the payload as JSON with the given HTTP status.
func successJSON[T any](w http.ResponseWriter, code int, payload T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := Response[T]{
		Code:   code,
		Status: http.StatusText(code),
		Result: payload,
		Error:  nil,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// todo: validate this log, to improve error handling
		log.Printf("failed to write JSON response: %v", err)
	}
}
