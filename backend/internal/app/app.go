package app

import (
	"log/slog"
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api"
	"github.com/clementd-tek/remote-buzzer/backend/internal/config"
)

type App struct {
	config config.Config

	logger *slog.Logger

	server *http.Server
}

func New(
	cfg config.Config,
	logger *slog.Logger,
) *App {

	router := api.NewRouter(
		logger,
	)

	server := &http.Server{

		Addr: ":" + cfg.Port,

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
