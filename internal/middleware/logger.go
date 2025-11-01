package middleware

import (
	"net/http"
	"strings"

	"github.com/brxyxn/go-logger"
)

// Logger is a middleware that logs the incoming requests.
func Logger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Initialize the logger with the desired configuration.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info().
				Str("method", strings.ToLower(r.Method)).
				Str("path", r.URL.Path).
				Send()
			next.ServeHTTP(w, r)
		})
	}
}
