package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Kaese72/huemie-lib/liberrors"
	"github.com/golang-jwt/jwt/v5"
)

// LoadPublicKeyFromFile reads a PKIX PEM-encoded RSA public key from disk.
func LoadPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key file: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key at %s is not an RSA public key", path)
	}
	return rsaPub, nil
}

// UseTokenMiddleware returns an HTTP middleware that validates RS256 bearer tokens
// issued by the authentication service. Requests whose path starts with any entry
// in skipPrefixes bypass validation entirely.
//
// The returned type is func(http.Handler) http.Handler, which is directly assignable
// to gorilla/mux's MiddlewareFunc without an explicit cast.
func UseTokenMiddleware(publicKey *rsa.PublicKey, skipPrefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, prefix := range skipPrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				liberrors.NewApiError(liberrors.Unauthorized, errors.New("missing bearer token")).WriteHTTP(w)
				return
			}
			tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return publicKey, nil
			})
			if err != nil {
				liberrors.NewApiError(liberrors.Unauthorized, errors.New("invalid or expired token")).WriteHTTP(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
