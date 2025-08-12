// middleware/idempotency.go — Redis-based Idempotency Middleware
package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

// IdempotencyMiddleware handles deduplication of incoming requests using Idempotency-Key header
func IdempotencyMiddleware(ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.Background()
			idKey := "idempotency:" + sha256Hex(key)

			// check Redis for existing response
			val, err := redisClient.Get(ctx, idKey).Bytes()
			if err == nil {
				log.Printf("♻️ Replaying cached response for key %s", key)
				w.Header().Set("X-Idempotent-Cache", "HIT")
				w.WriteHeader(http.StatusOK)
				w.Write(val)
				return
			}

			// capture response
			recorder := &responseRecorder{ResponseWriter: w, buf: new(bytes.Buffer)}
			next.ServeHTTP(recorder, r)

			// store in Redis
			_ = redisClient.Set(ctx, idKey, recorder.buf.Bytes(), ttl).Err()

			// write original response
			w.Header().Set("X-Idempotent-Cache", "MISS")
			w.WriteHeader(recorder.status)
			w.Write(recorder.buf.Bytes())
		})
	}
}

// helper to hash the key
func sha256Hex(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// responseRecorder captures response output
type responseRecorder struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	r.buf.Write(p)
	return r.ResponseWriter.Write(p)
}
