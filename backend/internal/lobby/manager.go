package lobby

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var ErrNotFound = errors.New("lobby not found")

type Manager struct {
	lobbies sync.Map
	cache   Cache
	logger  *slog.Logger
}

// NewManager creates a Manager. cache may be nil, in which case the
// manager runs purely in-memory (no persistence across restarts, no
// sharing across instances).
func NewManager(cache Cache, logger *slog.Logger) *Manager {
	return &Manager{
		cache:  cache,
		logger: logger,
	}
}

func (m *Manager) Add(lobby *Lobby) {
	m.lobbies.Store(
		lobby.ID,
		lobby,
	)

	m.saveToCache(lobby)
}

func (m *Manager) Get(id string) (*Lobby, error) {
	value, ok := m.lobbies.Load(id)

	if !ok {
		return nil, ErrNotFound
	}

	return value.(*Lobby), nil
}

func (m *Manager) Delete(id string) {
	m.lobbies.Delete(id)

	if m.cache == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.cache.Delete(ctx, id); err != nil && m.logger != nil {
			m.logger.Warn("cache: failed to delete lobby", "lobby_id", id, "error", err)
		}
	}()
}

func (m *Manager) List() []*Lobby {
	result := make([]*Lobby, 0)

	m.lobbies.Range(
		func(key, value any) bool {
			result = append(
				result,
				value.(*Lobby),
			)

			return true
		},
	)

	return result
}

// Touch re-persists a lobby's current snapshot to the cache. Call this
// after any mutation (join, ready, open, buzz) so the cached directory
// stays in sync with in-memory state.
func (m *Manager) Touch(lobby *Lobby) {
	m.saveToCache(lobby)
}

func (m *Manager) saveToCache(lobby *Lobby) {
	if m.cache == nil {
		return
	}

	snapshot := lobby.Snapshot()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := m.cache.Save(ctx, snapshot); err != nil && m.logger != nil {
			m.logger.Warn("cache: failed to save lobby", "lobby_id", snapshot.ID, "error", err)
		}
	}()
}

// Hydrate loads every lobby directory entry from the cache into memory.
// Call this once at startup, before serving traffic, so lobbies survive a
// backend restart. Player lists are not restored (see Restore).
func (m *Manager) Hydrate(ctx context.Context) error {
	if m.cache == nil {
		return nil
	}

	snapshots, err := m.cache.List(ctx)

	if err != nil {
		return err
	}

	for _, snapshot := range snapshots {
		m.lobbies.Store(
			snapshot.ID,
			Restore(snapshot),
		)
	}

	if m.logger != nil {
		m.logger.Info("cache: hydrated lobbies", "count", len(snapshots))
	}

	return nil
}

// StartCleanup runs a background loop that evicts lobbies which have had
// no activity for longer than ttl. It stops when ctx is cancelled.
func (m *Manager) StartCleanup(ctx context.Context, ttl time.Duration, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.evictStale(ttl)
			}
		}
	}()
}

func (m *Manager) evictStale(ttl time.Duration) {
	for _, l := range m.List() {
		if !l.IsStale(ttl) {
			continue
		}

		if m.logger != nil {
			m.logger.Info("cleanup: evicting stale lobby", "lobby_id", l.ID)
		}

		m.Delete(l.ID)
	}
}
