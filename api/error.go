package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Error writes an error message as a JSON response.
func Error(w http.ResponseWriter, code int, error ErrorResponse) {
	errorJSON(w, code, error)
}

// errorJSON writes the payload as JSON with the given HTTP status.
func errorJSON(w http.ResponseWriter, code int, payload ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := Response[any]{
		Code:   code,
		Status: http.StatusText(code),
		Error:  &payload,
		Result: nil,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}
