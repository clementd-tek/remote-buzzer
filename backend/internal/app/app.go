package app

import (
	"log/slog"
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api"
	"github.com/clementd-tek/remote-buzzer/backend/internal/config"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/clementd-tek/remote-buzzer/backend/internal/ws"
)

type App struct {
	config config.Config

	logger *slog.Logger

	server *http.Server
}

func New(cfg config.Config, logger *slog.Logger) *App {
	manager := lobby.NewManager()

	lobbyService := lobby.NewService(
		manager,
	)

	hub := ws.NewHub(logger)

	router := api.NewRouter(
		logger,
		lobbyService,
		hub,
	)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		config: cfg,
		logger: logger,
		server: server,
	}
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}
