package ratelimiter

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
	peakRate   int
}

func NewRateLimiter(cfg *config.Config) *RateLimiter {
	return &RateLimiter{
		requests:   make([]time.Time, 0, cfg.RateLimit+cfg.RateBuffer),
		rateLimit:  cfg.RateLimit,
		rateBuffer: cfg.RateBuffer,
		timeFrame:  cfg.TimeFrame,
		peakRate:   0,
	}
}

func (rl *RateLimiter) Add() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.requests = append(rl.requests, now)

	rate := rl.getRate()
	rl.updatePeakRate(rate)
}

func (rl *RateLimiter) Allow() (int, bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	curRate := rl.getRate()
	maxReqsToLock := rl.rateLimit + rl.rateBuffer

	if !rl.locked && curRate == maxReqsToLock {
		rl.locked = true
	}

	if rl.locked && curRate < rl.rateLimit {
		rl.locked = false
	}

	if rl.locked {
		return curRate, false
	}

	return curRate, true
}

func (rl *RateLimiter) GetLimit() time.Duration {
	return rl.timeFrame
}

func (rl *RateLimiter) GetPeakRate() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.peakRate
}

func (rl *RateLimiter) ResetPeakRate() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	lastPeakRate := rl.peakRate
	rl.peakRate = 0

	return lastPeakRate
}

func (rl *RateLimiter) getRate() int {
	now := time.Now()

	for len(rl.requests) > 0 && now.Sub(rl.requests[0]) > rl.timeFrame {
		rl.requests = rl.requests[1:]
	}

	return len(rl.requests)
}

func (rl *RateLimiter) updatePeakRate(rate int) {
	if rate > rl.peakRate {
		rl.peakRate = rate
	}
}
