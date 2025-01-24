package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"url-shortner/internal/config"
	"url-shortner/internal/http-server/handlers/redirect"
	"url-shortner/internal/http-server/handlers/url/deleteUrl"
	"url-shortner/internal/http-server/handlers/url/save"
	mwLimiter "url-shortner/internal/http-server/middleware/rateLimiter"
	rl "url-shortner/internal/rateLimiter"
)

type UrlOperator interface {
	save.URLSaver
	redirect.URLGetter
	deleteUrl.URLDeleter
}

func InitRoutes(cfg *config.Config, log *slog.Logger, storage UrlOperator, router *chi.Mux, rateLimiter *rl.RateLimiter) {
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Use(mwLimiter.New(log, rateLimiter))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", deleteUrl.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))
}
