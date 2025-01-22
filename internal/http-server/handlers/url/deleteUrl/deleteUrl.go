package deleteUrl

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortner/internal/api/response"
	"url-shortner/internal/storage"
)

type URLDeleter interface {
	DeleteURL(alias string) error
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.deleteUrl.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request-id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("missed 'alias' param", slog.Attr{Key: "error", Value: slog.StringValue("missed 'alias' param")})

			render.JSON(w, r, resp.Error("missed 'alias' param"))

			return
		}

		log.Info("got alias param", slog.String("alias", alias))

		_, err := urlDeleter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("url to delete not found", slog.String("alias", alias))

			render.JSON(w, r, resp.Error("url to delete not found"))

			return
		}

		err = urlDeleter.DeleteURL(alias)
		if err != nil {
			log.Error("failed to deleteUrl url", slog.String("error", err.Error()))

			render.JSON(w, r, resp.Error("failed to deleteUrl url"))

			return
		}

		log.Info("deleted url", slog.String("alias", alias))

		responseOK(w, r)
	})
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, resp.OK())
}
