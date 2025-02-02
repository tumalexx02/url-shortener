package stats

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortner/internal/api/response"
	"url-shortner/internal/models/database"
	ratelimiter "url-shortner/internal/rate-limiter"
	"url-shortner/internal/stats"
)

type StatisticGetter interface {
	GetStats() (database.Statistic, error)
}

type Response struct {
	resp.Response
	Statistic stats.Statistic `json:"stats,omitempty"`
}

func New(log *slog.Logger, statsGetter StatisticGetter, rateLimiter *ratelimiter.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.stats.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		dbStats, err := statsGetter.GetStats()
		if err != nil {
			log.Error("failed to get dbStats", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})

			render.JSON(w, r, resp.Error("failed to get dbStats"))

			return
		}

		var leaders []stats.ResourceInfo
		err = json.Unmarshal(dbStats.LeadersJSON, &leaders)
		if err != nil {
			log.Error("failed to unmarshal leaders", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		}

		log.Info("got leaders", slog.String("leaders", fmt.Sprintf("%v", dbStats.LeadersJSON)))

		statistic := stats.NewStatistic(dbStats.TotalURLCount, rateLimiter.GetRate(), dbStats.DayPeak, leaders)

		log.Info("got statistic", slog.String("stats", fmt.Sprintf("%+v", dbStats)))
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
