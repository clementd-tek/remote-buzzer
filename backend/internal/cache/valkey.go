// Package cache provides a Valkey (Redis-protocol compatible) backed
// implementation of lobby.Cache. It acts as a directory/cache for lobby
// snapshots so the lobby list survives a backend restart and can, in the
// future, be shared across multiple backend instances sitting behind the
// load balancer. It is intentionally not the source of truth for a live
// round: in-flight game state (websocket connections, exact player set)
// stays in the process that's actually running the round.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/redis/go-redis/v9"
)

// indexKey holds the set of every lobby id currently cached, so List
// doesn't need to scan the whole keyspace.
const indexKey = "lobbies:index"

// entryTTL bounds how long a cached lobby entry can live without being
// refreshed (via Save). This is a safety net in addition to the
// application-level cleanup in lobby.Manager.StartCleanup.
const entryTTL = 12 * time.Hour

type Store struct {
	client *redis.Client
}

// Config holds the connection settings for Valkey/Redis.
type Config struct {
	Addr     string
	Password string
	DB       int
}

// New creates a Store and verifies connectivity with a short-lived ping.
// Callers should treat a non-nil error as "run without a cache" rather
// than a fatal startup error, so local development without Valkey running
// still works.
func New(ctx context.Context, cfg Config) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("valkey: ping failed: %w", err)
	}

	return &Store{client: client}, nil
}

func (s *Store) Close() error {
	return s.client.Close()
}

func lobbyKey(id string) string {
	return "lobby:" + id
}

func (s *Store) Save(ctx context.Context, snapshot lobby.LobbySnapshot) error {
	payload, err := json.Marshal(snapshot)

	if err != nil {
		return err
	}

	pipe := s.client.TxPipeline()
	pipe.Set(ctx, lobbyKey(snapshot.ID), payload, entryTTL)
	pipe.SAdd(ctx, indexKey, snapshot.ID)

	_, err = pipe.Exec(ctx)

	return err
}

func (s *Store) Delete(ctx context.Context, id string) error {
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, lobbyKey(id))
	pipe.SRem(ctx, indexKey, id)

	_, err := pipe.Exec(ctx)

	return err
}

func (s *Store) List(ctx context.Context) ([]lobby.LobbySnapshot, error) {
	ids, err := s.client.SMembers(ctx, indexKey).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	snapshots := make([]lobby.LobbySnapshot, 0, len(ids))

	for _, id := range ids {
		payload, err := s.client.Get(ctx, lobbyKey(id)).Result()

		if errors.Is(err, redis.Nil) {
			// Entry expired (TTL) but the index still references it;
			// drop it from the index and move on.
			s.client.SRem(ctx, indexKey, id)
			continue
		}

		if err != nil {
			return nil, err
		}

		var snapshot lobby.LobbySnapshot

		if err := json.Unmarshal([]byte(payload), &snapshot); err != nil {
			continue
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}
