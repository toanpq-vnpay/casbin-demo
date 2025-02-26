package middlewares

import (
	"fmt"
	"net/http"

	"casbin-demo/models"

	"github.com/casbin/casbin/v2"
)

func Authorize(e *casbin.Enforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get username from JWT context
			claims, ok := r.Context().Value(ClaimsKey).(*models.Claims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			fmt.Println("username", claims.Username, ", path", r.URL.Path, ", method", r.Method)

			ok, err := e.Enforce(claims.Username, r.URL.Path, r.Method)
			if err != nil {
				http.Error(w, "Authorization error", http.StatusInternalServerError)
				return
			}

			if !ok {
				http.Error(w, "Permission denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
