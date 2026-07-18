package app

import (
	"log/slog"
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api"
	"github.com/clementd-tek/remote-buzzer/backend/internal/config"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
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

	router := api.NewRouter(
		logger,
		lobbyService,
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
