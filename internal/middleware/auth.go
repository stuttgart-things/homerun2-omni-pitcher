package middleware

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	homerun "github.com/stuttgart-things/homerun-library/v4"
)

// TokenAuthMiddleware validates bearer token authentication
func TokenAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithAuthError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check if the Authorization header has the Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithAuthError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]
		expectedToken := homerun.GetEnv("AUTH_TOKEN", "")

		if expectedToken == "" {
			respondWithAuthError(w, http.StatusInternalServerError, "Server authentication not configured")
			return
		}

		// Constant-time compare: a byte-wise != leaks how much of the secret
		// matched via timing. Mirrors hmac.Equal on the GitHub webhook path.
		if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			respondWithAuthError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Token is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	}
}

func respondWithAuthError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = fmt.Fprintf(w, `{"status":"error","message":"%s"}`, message)
}
