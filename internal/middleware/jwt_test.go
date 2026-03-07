package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// testJWKSServer creates a test HTTP server that serves a JWKS with the given RSA public key.
func testJWKSServer(t *testing.T, key *rsa.PublicKey) *httptest.Server {
	t.Helper()

	n := key.N.Bytes()
	e := big.NewInt(int64(key.E)).Bytes()

	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": "test-key-1",
				"n":   encodeBase64URL(n),
				"e":   encodeBase64URL(e),
			},
		},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
}

func encodeBase64URL(data []byte) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	result := make([]byte, 0, (len(data)*4+2)/3)
	for i := 0; i < len(data); i += 3 {
		var val uint32
		remaining := len(data) - i
		if remaining >= 3 {
			val = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
			result = append(result, alphabet[val>>18&0x3F], alphabet[val>>12&0x3F], alphabet[val>>6&0x3F], alphabet[val&0x3F])
		} else if remaining == 2 {
			val = uint32(data[i])<<16 | uint32(data[i+1])<<8
			result = append(result, alphabet[val>>18&0x3F], alphabet[val>>12&0x3F], alphabet[val>>6&0x3F])
		} else {
			val = uint32(data[i]) << 16
			result = append(result, alphabet[val>>18&0x3F], alphabet[val>>12&0x3F])
		}
	}
	return string(result)
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	srv := testJWKSServer(t, &privateKey.PublicKey)
	defer srv.Close()

	mw, err := NewJWTAuthMiddleware(JWTConfig{JWKSURL: srv.URL})
	if err != nil {
		t.Fatalf("failed to create JWT middleware: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "user1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "test-key-1"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/pitch", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	srv := testJWKSServer(t, &privateKey.PublicKey)
	defer srv.Close()

	mw, err := NewJWTAuthMiddleware(JWTConfig{JWKSURL: srv.URL})
	if err != nil {
		t.Fatalf("failed to create JWT middleware: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "user1",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	token.Header["kid"] = "test-key-1"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/pitch", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	srv := testJWKSServer(t, &privateKey.PublicKey)
	defer srv.Close()

	mw, err := NewJWTAuthMiddleware(JWTConfig{JWKSURL: srv.URL})
	if err != nil {
		t.Fatalf("failed to create JWT middleware: %v", err)
	}

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/pitch", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}
