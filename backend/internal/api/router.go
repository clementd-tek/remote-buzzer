package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	logger *slog.Logger,
) http.Handler {

	r := chi.NewRouter()

	r.Get(
		"/health",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			response := map[string]string{
				"status": "ok",
			}

			w.Header().
				Set(
					"Content-Type",
					"application/json",
				)

			json.NewEncoder(w).
				Encode(response)
		},
	)

	r.Get(
		"/",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			w.Write(
				[]byte(
					"remote buzzer server",
				),
			)
		},
	)

	return r
}
