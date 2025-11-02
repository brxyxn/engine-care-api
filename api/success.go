package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Success writes an success message as a JSON response.
func Success(w http.ResponseWriter, code int, messages interface{}) {
	successJSON(w, code, messages)
}

// successJSON writes the payload as JSON with the given HTTP status.
func successJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ResponseStruct{
		Code:   code,
		Status: http.StatusText(code),
		Data:   payload,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// todo: validate this log, to improve error handling
		log.Printf("failed to write JSON response: %v", err)
	}
}
