package main

import (
	"log/slog"
	"os"
	"url-shortner/internal/app"
	"url-shortner/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoadConfig()

	log := initLogger(cfg.Env)

	application, err := app.New(cfg, log)
	if err != nil {
		log.Error("failed to init app", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		os.Exit(1)
	}

	err = application.Start()
	if err != nil {
		log.Error("failed to start server")
	}

	log.Info("shutting down server")
}

func initLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
