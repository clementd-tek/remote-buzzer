package main

import (
	"log/slog"
	"os"

	"github.com/clementd-tek/remote-buzzer/backend/internal/app"
	"github.com/clementd-tek/remote-buzzer/backend/internal/config"
)

func main() {

	cfg := config.Load()

	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			nil,
		),
	)

	application := app.New(
		cfg,
		logger,
	)

	logger.Info(
		"starting server",
		"port",
		cfg.Port,
	)

	if err := application.Run(); err != nil {
		logger.Error(
			"server stopped",
			"error",
			err,
		)

		os.Exit(1)
	}
}
