package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"url-shortner/internal/models/database"
)

func (a *App) StartPeakRateResetJob(ctx context.Context) error {
	const op = "app.startDailyJobs"

	loc, err := time.LoadLocation(a.cfg.Location)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("starting daily peak reset", slog.String("location", a.cfg.Location))

	go func() {
		for {
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
			duration := time.Until(next)

			timer := time.NewTimer(duration)

			select {
			case <-ctx.Done():
				timer.Stop()
				a.log.Info("stopping daily peak reset")
				return
			case <-timer.C:
				lastPeakRate := a.rateLimiter.GetPeakRate()
				a.rateLimiter.ResetPeakRate()
				err := a.storage.ResetPeakRate()
				if err != nil {
					a.log.Error("failed resetting peak rate", slog.String("error", err.Error()))
					continue
				}
				a.log.Info("peak rate reset completed", slog.Int("last-peak-rate", lastPeakRate))
			}
		}
	}()

	return nil
}

func (a *App) StartAnalyticsJob(ctx context.Context) error {
	const op = "app.startAnalyticsJob"

	loc, err := time.LoadLocation(a.cfg.Location)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.initStats()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("starting analytics job", slog.String("location", a.cfg.Location))

	go func() {
		for {
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, now.Second(), 0, loc)
			duration := time.Until(next)

			timer := time.NewTimer(duration)

			select {
			case <-ctx.Done():
				timer.Stop()
				a.log.Info("stopping daily peak reset")
				return
			case <-timer.C:
				err := a.updateStats()
				if err != nil {
					a.log.Error("failed to update stats", slog.String("error", err.Error()))
					continue
				}
				now := time.Now()
				a.log.Info("analytics updated", slog.Time("time", now))
			}
		}
	}()

	return nil
}

func (a *App) initStats() error {
	const op = "app.initStats"

	existingPeakRate, err := a.storage.GetLastPeakRate()
	if err != nil {
		a.log.Error("failed getting last peak rate", slog.String("error", err.Error()))
	}

	a.log.Info("last peak rate got", slog.Int("day_peak", existingPeakRate.DayPeak), slog.Time("last_updated", existingPeakRate.LastUpdate), slog.Bool("is_yesterday", isYesterday(existingPeakRate.LastUpdate)))

	if existingPeakRate.DayPeak != 0 && !isYesterday(existingPeakRate.LastUpdate) {
		a.rateLimiter.SetPeakRate(existingPeakRate.DayPeak)
	}

	err = a.updateStats()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) updateStats() error {
	const op = "app.updateStats"

	totalUrl, err := a.storage.GetURLCount()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	leaders, err := a.storage.GetResourcesLeaders()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	peakRate := a.rateLimiter.GetPeakRate()

	leadersJSON, err := json.Marshal(leaders)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	statistic := database.NewStatistic(totalUrl, peakRate, leadersJSON)
	err = a.storage.UpdateStats(statistic)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func isYesterday(t time.Time) bool {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}
