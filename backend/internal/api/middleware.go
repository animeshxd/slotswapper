package api

import (
	"context"
	"net/http"
	"strings"

	"slotswapper/internal/crypto"
)

type contextKey string

const (
	userIDContextKey contextKey = "userID"
)

// AuthMiddleware is a middleware to authenticate requests using JWT from a cookie or Bearer token.
func AuthMiddleware(jwtManager crypto.JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// 1. Try to get the token from the cookie first
			cookie, err := r.Cookie("access_token")
			if err == nil {
				tokenString = cookie.Value
			}

			// 2. If no cookie, fall back to the Authorization header
			if tokenString == "" {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					http.Error(w, "Authorization required", http.StatusUnauthorized)
					return
				}

				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
					return
				}
				tokenString = parts[1]
			}

			if tokenString == "" {
				http.Error(w, "Token not found", http.StatusUnauthorized)
				return
			}

			userID, err := jwtManager.Verify(tokenString)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext extracts the user ID from the request context.
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}
