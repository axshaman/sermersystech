package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
    "bytes"
    "io/ioutil"
	"github.com/redis/go-redis/v9"
)

// RateLimitOptions задаёт параметры лимитирования
type RateLimitOptions struct {
	Limit     int                              // сколько попыток разрешено
	Window    time.Duration                    // временное окно
	KeyFunc   func(r *http.Request) string     // функция генерации уникального ключа
	ActionTag string                           // логическая метка (например: "topup")
}

// RateLimitMiddleware — middleware для ограничения частоты запросов
func RateLimitMiddleware(opts RateLimitOptions) func(http.Handler) http.Handler {
    if redisClient == nil {
        panic("Redis client not initialized! Call SetRedisClient() first.")
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := context.Background()
            key := fmt.Sprintf("ratelimit:%s", opts.KeyFunc(r))

            count, err := redisClient.Incr(ctx, key).Result()
            if err != nil {
                log.Printf("❌ Redis INCR error for key %s: %v", key, err)
                http.Error(w, "Rate limit error", http.StatusInternalServerError)
                return
            }

            if count == 1 {
                redisClient.Expire(ctx, key, opts.Window)
            }

            if int(count) > opts.Limit {
                ip := getIP(r)
                logKey := fmt.Sprintf("ratelimitlog:%s", opts.KeyFunc(r))
                body, _ := ioutil.ReadAll(r.Body)
                r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

                entry := fmt.Sprintf("[%s] %s %s from %s, body: %s",
                    time.Now().Format("2006-01-02 15:04:05"),
                    r.Method,
                    r.URL.Path,
                    ip,
                    string(body),
                )

                redisClient.LPush(ctx, logKey, entry)
                redisClient.LTrim(ctx, logKey, 0, 99)
                redisClient.Expire(ctx, logKey, 24*time.Hour)

                w.Header().Set("Retry-After", fmt.Sprintf("%.0f", opts.Window.Seconds()))
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// getIP извлекает реальный IP клиента из заголовков или RemoteAddr
func getIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // fallback
	}
	return ip
}