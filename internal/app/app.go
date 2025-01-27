package app

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	cr "github.com/robfig/cron/v3"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"url-shortner/internal/config"
	mwLogger "url-shortner/internal/http-server/middleware/logger"
	"url-shortner/internal/jobs"
	"url-shortner/internal/jobs/reset-peak-rate"
	rl "url-shortner/internal/rate-limiter"
	"url-shortner/internal/routes"
	"url-shortner/internal/storage/postgres"
)

type App struct {
	cnf     *config.Config
	log     *slog.Logger
	storage *postgres.Storage
	router  *chi.Mux
	rl      *rl.RateLimiter
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

	resetPeakJob := reset_peak_rate.New(a.log, a.rl)
	err := a.startJobsScheduler("0 0 * * *", resetPeakJob)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("starting server", slog.String("address", a.cnf.Address))

	srv := &http.Server{
		Addr:         a.cnf.Address,
		Handler:      a.router,
		ReadTimeout:  a.cnf.HTTPServer.Timeout,
		WriteTimeout: a.cnf.HTTPServer.Timeout,
		IdleTimeout:  a.cnf.HTTPServer.IdleTimeout,
	}

	return srv.ListenAndServe()
}

func (a *App) startJobsScheduler(cronPattern string, jobs ...*jobs.Job) error {
	const op = "app.startJobsScheduler"

	loc, err := time.LoadLocation(a.cnf.Location)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	c := cr.New(cr.WithLocation(loc))

	for _, job := range jobs {
		_, err = c.AddJob(cronPattern, job)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		a.log.Info("job added", slog.String("job", job.Name))
	}

	c.Start()

	a.log.Info("jobs scheduler enabled", slog.String("jobs", strconv.Itoa(len(jobs))))

	return nil
}
