package middlewares

import (
	"context"
	"fmt"
	"os"
	"strings"

	"net/http"

	"casbin-demo/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

const (
	bearerSchema = "Bearer "
)

var (
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if it's a Bearer token
	if strings.HasPrefix(authHeader, bearerSchema) {
		return strings.TrimPrefix(authHeader, bearerSchema)
	}

	// Return the raw header value if no Bearer prefix
	return authHeader
}

func Authenticate() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			claims := &models.Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
