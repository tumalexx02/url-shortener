package ratelimiter

import (
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	resp "url-shortner/internal/api/response"
	rl "url-shortner/internal/rate-limiter"
)

func New(log *slog.Logger, rateLimiter *rl.RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			slog.String("component", "middleware/rate-limiter"),
		)

		log.Info("rate-limiter middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			if rate, allow := rateLimiter.Allow(); !allow {
				log.Info("rate-limiter exceeded", slog.Int("rate", rate))
				rateLimitExceeded(w, r, int(rateLimiter.GetLimit().Seconds()))
				return
			}

			next.ServeHTTP(w, r)

			rateLimiter.Add()
		}

		return http.HandlerFunc(fn)
	}
}

func rateLimitExceeded(w http.ResponseWriter, r *http.Request, seconds int) {
	w.Header().Add("Retry-After", strconv.Itoa(seconds))
	w.WriteHeader(http.StatusInternalServerError)
	render.JSON(w, r, resp.Unavailable())
}
