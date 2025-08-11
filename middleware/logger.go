package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger logs HTTP method, path, status, and execution time
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, ww.statusCode, time.Since(start))
	})
}

// responseWriter allows capturing the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}