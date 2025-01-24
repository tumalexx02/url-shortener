package mwRateLimiter

import (
	"log/slog"
	"net/http"
	"strconv"
	rl "url-shortner/internal/rateLimiter"
)

func New(log *slog.Logger, rateLimiter *rl.RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/rateLimiter"),
		)

		log.Info("rateLimiter middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.Allow() {
				rateLimitExceeded(w, int(rateLimiter.GetLimit().Seconds()))
				return
			}

			next.ServeHTTP(w, r)

			rateLimiter.Add()
		}

		return http.HandlerFunc(fn)
	}
}

func rateLimitExceeded(w http.ResponseWriter, seconds int) {
	w.Header().Add("Retry-After", strconv.Itoa(seconds))
	w.WriteHeader(http.StatusInternalServerError)
}
