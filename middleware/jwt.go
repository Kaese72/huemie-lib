package middleware

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Kaese72/huemie-lib/liberrors"
	"github.com/golang-jwt/jwt/v5"
)

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
