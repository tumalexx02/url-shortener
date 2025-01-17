package main

import (
	"log/slog"
	"os"
	"url-shortner/internal/config"
	"url-shortner/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoadConfig()

	log := initLogger(cfg.Env)

	log.Info("starting server", slog.String("env", cfg.Env))

	storage, err := postgres.New(cfg.PostgresConfig)
	if err != nil {
		log.Error("failed to init storage", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		os.Exit(1)
	}

	_ = storage

	// TODO: router init

	// TODO: run server
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
