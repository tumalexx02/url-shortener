package redirect

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

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirect.New"

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

		url, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Info("url not found", slog.String("alias", alias))

			render.JSON(w, r, resp.Error("url not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})

			render.JSON(w, r, resp.Error("failed to get url"))

			return
		}

		log.Info("got url", slog.String("alias", alias), slog.String("url", url))

		url = "https://" + url
		http.Redirect(w, r, url, http.StatusFound)
	}
}
