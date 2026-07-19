package api

import (
	"log/slog"
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api/handlers"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/clementd-tek/remote-buzzer/backend/internal/originpolicy"
	"github.com/clementd-tek/remote-buzzer/backend/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(logger *slog.Logger, lobbyService *lobby.Service, hub *ws.Hub, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.Logger,
		middleware.Recoverer,
	)

	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}

	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return originpolicy.IsLocal(origin) || len(allowed) == 0 || allowed[origin]
		},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	lobbyHandler := handlers.NewLobbyHandler(
		lobbyService,
		hub,
	)

	lobbyWSHandler := handlers.NewLobbyWSHandler(
		lobbyService,
		hub,
	)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route(
		"/api",
		func(r chi.Router) {

			r.Post(
				"/lobbies",
				lobbyHandler.Create,
			)

			r.Get(
				"/lobbies",
				lobbyHandler.List,
			)

			r.Get(
				"/lobbies/{id}",
				lobbyHandler.Get,
			)

			r.Post(
				"/lobbies/{id}/join",
				lobbyHandler.Join,
			)

			r.Get(
				"/lobbies/{id}/ws",
				lobbyWSHandler.Serve,
			)
		},
	)

	return r
}
