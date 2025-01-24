package rateLimiter

import (
	"sync"
	"time"
	"url-shortner/internal/config"
)

type RateLimiter struct {
	mu         sync.RWMutex
	requests   []time.Time
	rateLimit  int
	rateBuffer int
	timeFrame  time.Duration
	locked     bool
	peakRate   PeakRate
}

type PeakRate struct {
	rate      int
	resetTime time.Time
}

func NewRateLimiter(cfg *config.Config) *RateLimiter {
	return &RateLimiter{
		requests:   make([]time.Time, 0, cfg.RateLimit+cfg.RateBuffer),
		rateLimit:  cfg.RateLimit,
		rateBuffer: cfg.RateBuffer,
		timeFrame:  cfg.TimeFrame,
	}
}

func (rl *RateLimiter) Add() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.requests = append(rl.requests, now)
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	curRate := rl.GetRate()
	maxReqsToLock := rl.rateLimit + rl.rateBuffer

	if !rl.locked && curRate == maxReqsToLock {
		rl.locked = true
	}

	if rl.locked && curRate < rl.rateLimit {
		rl.locked = false
	}

	if rl.locked {
		return false
	}

	return true
}

func (rl *RateLimiter) GetRate() int {
	now := time.Now()

	for len(rl.requests) > 0 && now.Sub(rl.requests[0]) > rl.timeFrame {
		rl.requests = rl.requests[1:]
	}

	return len(rl.requests)
}

func (rl *RateLimiter) GetLimit() time.Duration {
	return rl.timeFrame
}

func (rl *RateLimiter) GetPeakRate() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	rl.checkPeakRate()

	return rl.peakRate.rate
}

func (rl *RateLimiter) updatePeakRate(rate int) {
	rl.checkPeakRate()

	if rate > rl.peakRate.rate {
		rl.peakRate.rate = rate
	}
}

func (rl *RateLimiter) checkPeakRate() {
	now := time.Now()

	if now.Sub(rl.peakRate.resetTime) >= 24*time.Hour {
		rl.peakRate.rate = 0
		rl.peakRate.resetTime = now
	}
}
