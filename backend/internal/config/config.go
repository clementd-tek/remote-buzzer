package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port string

	// FrontendOrigins is the comma-separated list of allowed CORS /
	// websocket origins for the React frontend.
	//
	// Configured via FRONTEND_ORIGINS in the environment (or .env).
	// localhost / 127.0.0.1 / ::1 are always permitted regardless of
	// this value (see originpolicy.IsLocal).
	//
	// Examples:
	//   FRONTEND_ORIGINS=https://buzzer.example.com
	//   FRONTEND_ORIGINS=https://buzzer.example.com, https://admin.example.com
	FrontendOrigins []string

	// ValkeyAddr enables the Valkey cache when non-empty (e.g.
	// "valkey:6379" in Docker Compose, "localhost:6379" locally).
	// Leave empty to run purely in-memory.
	ValkeyAddr     string
	ValkeyPassword string
	ValkeyDB       int

	// LobbyTTL is how long a lobby can go without activity before it is
	// evicted by the cleanup loop. CleanupInterval is how often that
	// loop runs.
	LobbyTTL        time.Duration
	CleanupInterval time.Duration

	// ShutdownTimeout bounds how long graceful shutdown waits for
	// in-flight requests and open websocket connections to drain.
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		FrontendOrigins: parseOrigins(getEnv("FRONTEND_ORIGINS", "")),
		ValkeyAddr:      getEnv("VALKEY_ADDR", ""),
		ValkeyPassword:  getEnv("VALKEY_PASSWORD", ""),
		ValkeyDB:        getEnvInt("VALKEY_DB", 0),
		LobbyTTL:        getEnvDuration("LOBBY_TTL", 6*time.Hour),
		CleanupInterval: getEnvDuration("LOBBY_CLEANUP_INTERVAL", 10*time.Minute),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

// parseOrigins splits a comma-separated list of origins, trimming
// whitespace and dropping empty entries.
//
//	"https://a.com, https://b.com" → ["https://a.com", "https://b.com"]
//	""                             → []
func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	return out
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)

	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)

	if err != nil {
		return fallback
	}

	return parsed
}
