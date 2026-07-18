package api

import (
	"log/slog"
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api/handlers"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/clementd-tek/remote-buzzer/backend/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logger *slog.Logger, lobbyService *lobby.Service, hub *ws.Hub) http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.Logger,
		middleware.Recoverer,
	)

	lobbyHandler := handlers.NewLobbyHandler(
		lobbyService,
		hub,
	)

	lobbyWSHandler := handlers.NewLobbyWSHandler(
		lobbyService,
		hub,
	)

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
