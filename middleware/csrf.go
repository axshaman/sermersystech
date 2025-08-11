package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

// GenerateCSRFToken creates a secure random token
func GenerateCSRFToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Println("❌ CSRF token generation failed:", err)
		return ""
	}
	return hex.EncodeToString(b)
}

// CSRF protects against cross-site request forgery attacks.
// It sets a cookie if it's missing and validates the token on unsafe methods.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil || csrfCookie.Value == "" {
			token := GenerateCSRFToken()
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    token,
				HttpOnly: false,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				Path:     "/",
				Expires:  time.Now().Add(1 * time.Hour),
			})
		}

		// Validate on unsafe methods only
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			header := r.Header.Get("X-CSRF-Token")
			if csrfCookie == nil || csrfCookie.Value != header {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
