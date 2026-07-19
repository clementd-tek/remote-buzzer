package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	application := app.New(
		cfg,
		logger,
	)

	logger.Info(
		"starting server",
		"port",
		cfg.Port,
	)

	if err := application.Run(ctx); err != nil {
		logger.Error(
			"server stopped",
			"error",
			err,
		)

		os.Exit(1)
	}
}
