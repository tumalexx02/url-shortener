package ratelimiter_test

import (
	"testing"
	"time"
	"url-shortner/internal/config"
	"url-shortner/internal/rate-limiter"
)

func TestRateLimiter_BlockWithBuffer(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(&config.Config{RateLimiter: config.RateLimiter{
		RateLimit:  2,
		RateBuffer: 3,
		TimeFrame:  300 * time.Millisecond,
	}})

	rl.Add() // 00:00:00
	rl.Add() // 00:00:00

	time.Sleep(200 * time.Millisecond)

	rl.Add() // 00:00:03
	rl.Add() // 00:00:03
	rl.Add() // 00:00:03

	_, allow := rl.Allow()
	if allow {
		t.Fatal("expected false, got true")
	}

	time.Sleep(200 * time.Millisecond)

	_, allow = rl.Allow()
	if allow {
		t.Fatal("expected false, got true")
	}

	time.Sleep(200 * time.Millisecond)

	_, allow = rl.Allow()
	if !allow {
		t.Fatal("expected true, got false")
	}
}

func TestRateLimiter_GetRate(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(&config.Config{RateLimiter: config.RateLimiter{
		RateLimit:  2,
		RateBuffer: 0,
		TimeFrame:  100 * time.Millisecond,
	}})

	rl.Add()
	rl.Add()

	time.Sleep(150 * time.Millisecond)

	if rate, _ := rl.Allow(); rate != 0 {
		t.Fatal("expected 0, got ", rate)
	}
}

func TestRateLimiter_GetPeakRate(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(&config.Config{RateLimiter: config.RateLimiter{
		RateLimit:  100,
		RateBuffer: 5,
		TimeFrame:  100 * time.Millisecond,
	}})

	for range 50 {
		rl.Add()
	}

	if peakRate := rl.GetPeakRate(); peakRate != 50 {
		t.Fatalf("expected peakRate 50, got %d", peakRate)
	}

	rl.Add()

	if peakRate := rl.GetPeakRate(); peakRate != 51 {
		t.Fatalf("expected peakRate 51, got %d", peakRate)
	}
}

func TestRateLimiter_ResetPeakRate(t *testing.T) {
	rl := ratelimiter.NewRateLimiter(&config.Config{RateLimiter: config.RateLimiter{
		RateLimit:  100,
		RateBuffer: 5,
		TimeFrame:  100 * time.Millisecond,
	}})

	for range 50 {
		rl.Add()
	}

	if peakRate := rl.GetPeakRate(); peakRate != 50 {
		t.Fatalf("expected peakRate 50, got %d", peakRate)
	}

	time.Sleep(150 * time.Millisecond)
	rl.ResetPeakRate()

	for range 20 {
		rl.Add()
	}

	if peakRate := rl.GetPeakRate(); peakRate != 20 {
		t.Fatalf("expected peakRate 20, got %d", peakRate)
	}
}
