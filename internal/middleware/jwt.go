package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds configuration for JWT validation.
type JWTConfig struct {
	JWKSURL  string // JWKS endpoint URL
	Issuer   string // Expected iss claim (optional)
	Audience string // Expected aud claim (optional)
}

// NewJWTAuthMiddleware creates a middleware that validates JWTs against a JWKS endpoint.
func NewJWTAuthMiddleware(cfg JWTConfig) (func(http.HandlerFunc) http.HandlerFunc, error) {
	jwks, err := keyfunc.NewDefault([]string{cfg.JWKSURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS keyfunc from %s: %w", cfg.JWKSURL, err)
	}

	slog.Info("jwt auth configured", "jwks_url", cfg.JWKSURL)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithAuthError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithAuthError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			parserOpts := []jwt.ParserOption{jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"})}
			if cfg.Issuer != "" {
				parserOpts = append(parserOpts, jwt.WithIssuer(cfg.Issuer))
			}
			if cfg.Audience != "" {
				parserOpts = append(parserOpts, jwt.WithAudience(cfg.Audience))
			}

			token, err := jwt.Parse(tokenString, jwks.KeyfuncCtx(context.Background()), parserOpts...)
			if err != nil || !token.Valid {
				slog.Debug("jwt validation failed", "error", err)
				respondWithAuthError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			next.ServeHTTP(w, r)
		}
	}, nil
}
