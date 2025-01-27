package reset_peak_rate

import (
	"log/slog"
	"time"
	"url-shortner/internal/jobs"
	"url-shortner/internal/rate-limiter"
)

type ResetJob struct {
	log *slog.Logger
	rl  *ratelimiter.RateLimiter
}

func New(log *slog.Logger, rl *ratelimiter.RateLimiter) *jobs.Job {
	return &jobs.Job{
		Name: "jobs/reset-peak-rate",
		Job:  &ResetJob{log, rl},
	}
}

func (j *ResetJob) Run() {
	now := time.Now()
	lastPeakRate := j.rl.ResetPeakRate()
	j.log.Info("reset peak rate", slog.Time("time", now), slog.Int("last_peak_rate", lastPeakRate))
}
