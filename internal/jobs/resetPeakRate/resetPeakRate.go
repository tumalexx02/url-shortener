package resetPeakRate

import (
	"log/slog"
	"time"
	"url-shortner/internal/jobs"
	"url-shortner/internal/rateLimiter"
)

type ResetJob struct {
	log *slog.Logger
	rl  *rateLimiter.RateLimiter
}

func New(log *slog.Logger, rl *rateLimiter.RateLimiter) *jobs.Job {
	return &jobs.Job{
		Name: "jobs/resetPeakRate",
		Job:  &ResetJob{log, rl},
	}
}

func (j *ResetJob) Run() {
	now := time.Now()
	lastPeakRate := j.rl.ResetPeakRate()
	j.log.Info("reset peak rate", slog.Time("time", now), slog.Int("last_peak_rate", lastPeakRate))
}
