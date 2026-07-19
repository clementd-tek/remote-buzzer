package lobby_test

import (
	"context"
	"sync"

	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
)

// fakeCache is a minimal in-memory stand-in for the Valkey-backed cache,
// used so lobby package tests don't need a real Valkey instance.
type fakeCache struct {
	mu    sync.Mutex
	items map[string]lobby.LobbySnapshot
}

func newFakeCache() *fakeCache {
	return &fakeCache{items: make(map[string]lobby.LobbySnapshot)}
}

func (c *fakeCache) Save(_ context.Context, snapshot lobby.LobbySnapshot) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[snapshot.ID] = snapshot

	return nil
}

func (c *fakeCache) Delete(_ context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, id)

	return nil
}

func (c *fakeCache) List(_ context.Context) ([]lobby.LobbySnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]lobby.LobbySnapshot, 0, len(c.items))

	for _, snapshot := range c.items {
		result = append(result, snapshot)
	}

	return result, nil
}

func (c *fakeCache) has(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.items[id]

	return ok
}
