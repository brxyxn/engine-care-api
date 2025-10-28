package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Error writes an error message as a JSON response.
func Error(w http.ResponseWriter, code int, message string) {
	errorJSON(w, code, message)
}

// errorJSON writes the payload as JSON with the given HTTP status.
func errorJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ResponseStruct{
		Code:   code,
		Status: http.StatusText(code),
		Error:  payload,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}
