package lobby_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// waitUntil polls fn until it returns true or the timeout elapses. Manager
// writes to the cache asynchronously, so tests need to wait rather than
// assert immediately.
func waitUntil(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if fn() {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("condition not met before timeout")
}

func TestManagerAddSyncsToCache(t *testing.T) {
	cache := newFakeCache()
	manager := lobby.NewManager(cache, discardLogger())

	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())
	manager.Add(l)

	waitUntil(t, time.Second, func() bool { return cache.has("abc") })
}

func TestManagerDeleteRemovesFromCache(t *testing.T) {
	cache := newFakeCache()
	manager := lobby.NewManager(cache, discardLogger())

	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())
	manager.Add(l)

	waitUntil(t, time.Second, func() bool { return cache.has("abc") })

	manager.Delete("abc")

	waitUntil(t, time.Second, func() bool { return !cache.has("abc") })

	if _, err := manager.Get("abc"); err != lobby.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestManagerHydrateRestoresDirectoryFromCache(t *testing.T) {
	cache := newFakeCache()

	seed := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())
	if err := cache.Save(context.Background(), seed.Snapshot()); err != nil {
		t.Fatal(err)
	}

	manager := lobby.NewManager(cache, discardLogger())

	if err := manager.Hydrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	restored, err := manager.Get("abc")

	if err != nil {
		t.Fatalf("expected lobby to be restored, got error: %v", err)
	}

	if restored.Name != "test" {
		t.Fatalf("expected restored name %q, got %q", "test", restored.Name)
	}
}

func TestManagerCleanupEvictsStaleLobbies(t *testing.T) {
	cache := newFakeCache()
	manager := lobby.NewManager(cache, discardLogger())

	manager.Add(lobby.New("stale", "old", "host", true, lobby.DefaultSettings()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A negative ttl means every lobby is immediately stale, so the very
	// first tick should evict it.
	manager.StartCleanup(ctx, -time.Second, 10*time.Millisecond)

	waitUntil(t, time.Second, func() bool {
		_, err := manager.Get("stale")
		return err == lobby.ErrNotFound
	})
}
