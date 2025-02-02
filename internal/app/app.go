package app

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"url-shortner/internal/config"
	mwLogger "url-shortner/internal/http-server/middleware/logger"
	rl "url-shortner/internal/rate-limiter"
	"url-shortner/internal/routes"
	"url-shortner/internal/storage/postgres"
)

type App struct {
	cfg         *config.Config
	log         *slog.Logger
	storage     *postgres.Storage
	router      *chi.Mux
	rateLimiter *rl.RateLimiter
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	const op = "app.New"

	storage, err := postgres.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	rateLimiter := rl.NewRateLimiter(cfg)

	routes.InitRoutes(cfg, log, storage, router, rateLimiter)

	return &App{
		cfg,
		log,
		storage,
		router,
		rateLimiter,
	}, nil
}

func (a *App) Start() error {
	const op = "app.Start"

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	err := a.StartPeakRateResetJob(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.StartAnalyticsJob(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("starting server", slog.String("address", a.cfg.Address))

	srv := &http.Server{
		Addr:         a.cfg.Address,
		Handler:      a.router,
		ReadTimeout:  a.cfg.HTTPServer.Timeout,
		WriteTimeout: a.cfg.HTTPServer.Timeout,
		IdleTimeout:  a.cfg.HTTPServer.IdleTimeout,
	}

	return srv.ListenAndServe()
}
