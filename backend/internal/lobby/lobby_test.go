package lobby_test

import (
	"errors"
	"testing"
	"time"

	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
)

func TestBuzzWinner(t *testing.T) {
	l := lobby.New(
		"abc",
		"test",
		"host",
		true,
	)

	err := l.AddPlayer(
		&lobby.Player{
			ID: "p1",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	// lobby becomes ready
	err = l.SetReady()

	if err != nil {
		t.Fatal(err)
	}

	// host opens buzzer
	err = l.OpenBuzz()

	if err != nil {
		t.Fatal(err)
	}

	result, err := l.Buzz("p1")

	if err != nil {
		t.Fatal(err)
	}

	if result.PlayerID != "p1" {
		t.Fatal("wrong winner")
	}
}

func TestBuzzRejectsUnknownPlayer(t *testing.T) {
	l := lobby.New("abc", "test", "host", true)

	if err := l.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
	}

	if err := l.SetReady(); err != nil {
		t.Fatal(err)
	}

	if err := l.OpenBuzz(); err != nil {
		t.Fatal(err)
	}

	if _, err := l.Buzz("ghost"); !errors.Is(err, lobby.ErrPlayerNotFound) {
		t.Fatalf("expected ErrPlayerNotFound, got %v", err)
	}
}

func TestBuzzRejectsSecondWinner(t *testing.T) {
	l := lobby.New("abc", "test", "host", true)

	for _, id := range []string{"p1", "p2"} {
		if err := l.AddPlayer(&lobby.Player{ID: id}); err != nil {
			t.Fatal(err)
		}
	}

	if err := l.SetReady(); err != nil {
		t.Fatal(err)
	}

	if err := l.OpenBuzz(); err != nil {
		t.Fatal(err)
	}

	if _, err := l.Buzz("p1"); err != nil {
		t.Fatal(err)
	}

	// Once someone buzzes, the lobby moves to Locked, so a second buzz
	// is rejected because the round is closed (ErrAlreadyBuzzed guards
	// the same-tick race, but State already short-circuits first).
	if _, err := l.Buzz("p2"); !errors.Is(err, lobby.ErrLobbyClosed) {
		t.Fatalf("expected ErrLobbyClosed, got %v", err)
	}
}

func TestJoinRejectedOnceRoundIsOpen(t *testing.T) {
	l := lobby.New("abc", "test", "host", true)

	if err := l.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
	}

	if err := l.SetReady(); err != nil {
		t.Fatal(err)
	}

	if err := l.OpenBuzz(); err != nil {
		t.Fatal(err)
	}

	if err := l.AddPlayer(&lobby.Player{ID: "late"}); !errors.Is(err, lobby.ErrRoundInProgress) {
		t.Fatalf("expected ErrRoundInProgress, got %v", err)
	}
}

func TestValidateName(t *testing.T) {
	if err := lobby.ValidateName(""); err == nil {
		t.Fatal("expected error for empty name")
	}

	if err := lobby.ValidateName("   "); err == nil {
		t.Fatal("expected error for blank name")
	}

	if err := lobby.ValidateName("Clément"); err != nil {
		t.Fatalf("expected valid name, got %v", err)
	}
}

func TestIsStale(t *testing.T) {
	l := lobby.New("abc", "test", "host", true)

	if l.IsStale(time.Hour) {
		t.Fatal("freshly created lobby should not be stale")
	}

	if !l.IsStale(-time.Second) {
		t.Fatal("lobby should be considered stale with a negative ttl")
	}
}

func TestRestorePreservesDirectoryFieldsOnly(t *testing.T) {
	original := lobby.New("abc", "test", "host", true)

	if err := original.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
	}

	snapshot := original.Snapshot()
	restored := lobby.Restore(snapshot)
	restoredSnapshot := restored.Snapshot()

	if restoredSnapshot.ID != snapshot.ID || restoredSnapshot.Name != snapshot.Name {
		t.Fatal("restore should preserve directory fields")
	}

	if restoredSnapshot.PlayerCount != 0 {
		t.Fatal("restore should not preserve session-only player list")
	}
}
