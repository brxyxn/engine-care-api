package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/brxyxn/engine-care-api/api"
	"github.com/brxyxn/engine-care-api/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// Auth is middleware that verifies the JWT token provided in the Authorization header.
func Auth(conf config.Config) func(http.Handler) http.Handler {
	jwtKey := conf.JwtSecret

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expect the header to be in the format "Bearer <token>".
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				api.Error(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				api.Error(w, http.StatusUnauthorized, "Invalid Authorization header format")
				return
			}
			tokenStr := parts[1]

			// Parse and validate the token.
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return []byte(jwtKey), nil
			})
			if err != nil || !token.Valid {
				api.Error(w, http.StatusUnauthorized, ErrInvalidToken.Error())
				return
			}

			// Token is valid; proceed with the request.
			next.ServeHTTP(w, r)
		})
	}
}
