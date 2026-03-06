package middleware

import (
	"net/http"
	"strings"

	homerun "github.com/stuttgart-things/homerun-library"
)

// TokenAuthMiddleware validates bearer token authentication
func TokenAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"status":"error","message":"Missing authorization header"}`, http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		// Check if the Authorization header has the Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"status":"error","message":"Invalid authorization header format"}`, http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		token := parts[1]
		expectedToken := homerun.GetEnv("AUTH_TOKEN", "")

		if expectedToken == "" {
			http.Error(w, `{"status":"error","message":"Server authentication not configured"}`, http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		if token != expectedToken {
			http.Error(w, `{"status":"error","message":"Invalid token"}`, http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		// Token is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	}
}
