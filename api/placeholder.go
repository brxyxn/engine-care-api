package api

import "net/http"

func Placeholder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Success[string](w, http.StatusOK, "This is a placeholder endpoint.")
	}
}
