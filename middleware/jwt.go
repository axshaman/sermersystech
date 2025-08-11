package middleware

import (
	"context"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v4"
	"log"
)

// ContextKey is used to store/retrieve values in context
type ContextKey string

const UserIDKey ContextKey = "user_id"

// JWTMiddleware returns a middleware that checks for a valid JWT in the Authorization header.
// If valid, the user ID is injected into the request context.
func JWTMiddleware(jwtKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenStr string

			// 1. From Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// 2. Fallback: from cookie
				cookie, err := r.Cookie("auth_token")
				if err != nil || cookie.Value == "" {
					http.Error(w, "Unauthorized JWTM1", http.StatusUnauthorized)
					return
				}
				tokenStr = cookie.Value
			}

			claims := &struct {
				UserID int `json:"user_id"`
				jwt.StandardClaims
			}{}
			log.Printf("🔐 [JWTMiddleware] using jwtKey: %s", jwtKey)

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized JWTM2", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}


// GetUserID retrieves user ID from request context.
// Returns -1 if not present.
func GetUserID(r *http.Request) int {
	id, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		return -1
	}
	return id
}