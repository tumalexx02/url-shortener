package app

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"url-shortner/internal/config"
	mwLogger "url-shortner/internal/http-server/middleware/logger"
	rl "url-shortner/internal/rateLimiter"
	"url-shortner/internal/routes"
	"url-shortner/internal/storage/postgres"
)

type App struct {
	Config      *config.Config
	Logger      *slog.Logger
	Storage     *postgres.Storage
	Router      *chi.Mux
	RateLimiter *rl.RateLimiter
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
	a.Logger.Info("starting server", slog.String("address", a.Config.Address))

	srv := &http.Server{
		Addr:         a.Config.Address,
		Handler:      a.Router,
		ReadTimeout:  a.Config.HTTPServer.Timeout,
		WriteTimeout: a.Config.HTTPServer.Timeout,
		IdleTimeout:  a.Config.HTTPServer.IdleTimeout,
	}

	return srv.ListenAndServe()
}
