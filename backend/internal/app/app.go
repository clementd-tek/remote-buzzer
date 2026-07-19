package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api"
	"github.com/clementd-tek/remote-buzzer/backend/internal/cache"
	"github.com/clementd-tek/remote-buzzer/backend/internal/config"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/clementd-tek/remote-buzzer/backend/internal/ws"
)

type App struct {
	config config.Config

	logger *slog.Logger

	server  *http.Server
	manager *lobby.Manager
	store   *cache.Store

	cleanupCancel context.CancelFunc
}

func New(cfg config.Config, logger *slog.Logger) *App {
	var (
		lobbyCache lobby.Cache
		store      *cache.Store
	)

	// A missing or unreachable Valkey is not fatal: the app falls back
	// to running purely in-memory, which is fine for local dev.
	if cfg.ValkeyAddr != "" {
		s, err := cache.New(context.Background(), cache.Config{
			Addr:     cfg.ValkeyAddr,
			Password: cfg.ValkeyPassword,
			DB:       cfg.ValkeyDB,
		})

		if err != nil {
			logger.Warn(
				"valkey unreachable, running without cache",
				"addr", cfg.ValkeyAddr,
				"error", err,
			)
		} else {
			store = s
			lobbyCache = s
			logger.Info("valkey cache connected", "addr", cfg.ValkeyAddr)
		}
	}

	manager := lobby.NewManager(lobbyCache, logger)

	if lobbyCache != nil {
		if err := manager.Hydrate(context.Background()); err != nil {
			logger.Warn("failed to hydrate lobbies from cache", "error", err)
		}
	}

	lobbyService := lobby.NewService(
		manager,
	)

	allowedOrigins := splitOrigins(cfg.FrontendOrigin)

	hub := ws.NewHub(logger, allowedOrigins)

	router := api.NewRouter(
		logger,
		lobbyService,
		hub,
		allowedOrigins,
	)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	cleanupCtx, cancel := context.WithCancel(context.Background())
	manager.StartCleanup(cleanupCtx, cfg.LobbyTTL, cfg.CleanupInterval)

	return &App{
		config:        cfg,
		logger:        logger,
		server:        server,
		manager:       manager,
		store:         store,
		cleanupCancel: cancel,
	}
}

// Run starts the HTTP server and blocks until ctx is cancelled (typically
// on SIGINT/SIGTERM), at which point it drains in-flight requests and
// open websocket connections before returning.
func (a *App) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		a.logger.Info("shutting down")
		return a.Shutdown()
	}
}

func (a *App) Shutdown() error {
	a.cleanupCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
	defer cancel()

	err := a.server.Shutdown(shutdownCtx)

	if a.store != nil {
		if closeErr := a.store.Close(); closeErr != nil && a.logger != nil {
			a.logger.Warn("failed to close valkey connection", "error", closeErr)
		}
	}

	return err
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)

		if p != "" {
			origins = append(origins, p)
		}
	}

	return origins
}
