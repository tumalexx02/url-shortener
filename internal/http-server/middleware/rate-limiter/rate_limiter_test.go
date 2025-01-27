package ratelimiter_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"url-shortner/internal/config"
	mwRateLimiter "url-shortner/internal/http-server/middleware/rate-limiter"
	"url-shortner/internal/rate-limiter"
)

func TestRateLimiterMiddleware_BlockRequests(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(&config.Config{RateLimiter: config.RateLimiter{
		RateLimit:  1,
		RateBuffer: 0,
		TimeFrame:  time.Second,
	}})

	logger := slog.Default()

	middleware := mwRateLimiter.New(logger, rl)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}
}

func TestRateLimiterMiddleware_AllowRequests(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(&config.Config{RateLimiter: config.RateLimiter{
		RateLimit:  2,
		RateBuffer: 1,
		TimeFrame:  time.Second,
	}})

	logger := slog.Default()

	middleware := mwRateLimiter.New(logger, rl)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
