package stats

import (
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"url-shortner/internal/stats"
)

type StatisticGetter interface {
	GetStats() (stats.Statistic, error)
}

func New(log *slog.Logger, statsGetter StatisticGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.stats.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

	}
}
