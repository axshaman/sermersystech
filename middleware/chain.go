package middleware

import "net/http"

// Chain applies a list of middleware in order (left to right)
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}