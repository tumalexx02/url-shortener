package stats

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortner/internal/api/response"
	"url-shortner/internal/stats"
)

type StatisticGetter interface {
	GetStats() (stats.Statistic, error)
}

type Response struct {
	resp.Response
	Statistic stats.Statistic `json:"stats,omitempty"`
}

func New(log *slog.Logger, statsGetter StatisticGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.stats.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		statistic, err := statsGetter.GetStats()
		if err != nil {
			log.Error("failed to get statistic", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})

			render.JSON(w, r, resp.Error("failed to get statistic"))

			return
		}

		log.Info("got statistic")
		responseOK(w, r, statistic)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, stats stats.Statistic) {
	render.JSON(w, r,
		Response{
			resp.OK(),
			stats,
		},
	)
}
