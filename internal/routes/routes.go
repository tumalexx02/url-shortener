package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"url-shortner/internal/config"
	"url-shortner/internal/http-server/handlers/redirect"
	"url-shortner/internal/http-server/handlers/url/delete"
	"url-shortner/internal/http-server/handlers/url/save"
	mwRateLimiter "url-shortner/internal/http-server/middleware/rate-limiter"
	rl "url-shortner/internal/rate-limiter"
)

type UrlOperator interface {
	save.URLSaver
	redirect.URLGetter
	delete.URLDeleter
}

func InitRoutes(cfg *config.Config, log *slog.Logger, storage UrlOperator, router *chi.Mux, rateLimiter *rl.RateLimiter) {
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Use(mwRateLimiter.New(log, rateLimiter))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))
}
