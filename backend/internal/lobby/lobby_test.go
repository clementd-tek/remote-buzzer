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
		lobby.DefaultSettings(),
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
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

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
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

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
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

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
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	if l.IsStale(time.Hour) {
		t.Fatal("freshly created lobby should not be stale")
	}

	if !l.IsStale(-time.Second) {
		t.Fatal("lobby should be considered stale with a negative ttl")
	}
}

func TestRestorePreservesDirectoryFieldsOnly(t *testing.T) {
	original := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

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

func TestStartCountdownThenOpen(t *testing.T) {
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	if err := l.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
	}

	if err := l.SetReady(); err != nil {
		t.Fatal(err)
	}

	endsAt := time.Now().Add(3 * time.Second)

	if err := l.StartCountdown(endsAt); err != nil {
		t.Fatal(err)
	}

	snapshot := l.Snapshot()

	if snapshot.State != lobby.StateCountdown {
		t.Fatalf("expected state %q, got %q", lobby.StateCountdown, snapshot.State)
	}

	if snapshot.CountdownEndsAt == nil || !snapshot.CountdownEndsAt.Equal(endsAt) {
		t.Fatal("expected CountdownEndsAt to be set to the requested end time")
	}

	// The buzzer opening is what actually fires once the countdown
	// elapses (driven by a timer in the ws handler); confirm the
	// domain layer allows that countdown->open transition.
	if err := l.OpenBuzz(); err != nil {
		t.Fatal(err)
	}

	snapshot = l.Snapshot()

	if snapshot.State != lobby.StateOpen {
		t.Fatalf("expected state %q, got %q", lobby.StateOpen, snapshot.State)
	}

	if snapshot.CountdownEndsAt != nil {
		t.Fatal("expected CountdownEndsAt to be cleared once open")
	}
}

func TestJoinRejectedDuringCountdown(t *testing.T) {
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	if err := l.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
	}

	if err := l.SetReady(); err != nil {
		t.Fatal(err)
	}

	if err := l.StartCountdown(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	if err := l.AddPlayer(&lobby.Player{ID: "late"}); !errors.Is(err, lobby.ErrRoundInProgress) {
		t.Fatalf("expected ErrRoundInProgress, got %v", err)
	}
}

func TestNextRoundAwardsPointsAndResets(t *testing.T) {
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	for _, p := range []struct{ id, name string }{{"p1", "Alice"}, {"p2", "Bob"}} {
		if err := l.AddPlayer(&lobby.Player{ID: p.id, Name: p.name}); err != nil {
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

	record, err := l.NextRound()

	if err != nil {
		t.Fatal(err)
	}

	if record == nil || record.WinnerID != "p1" || record.WinnerName != "Alice" || record.Points != lobby.DefaultPointsPerRound {
		t.Fatalf("unexpected round record: %+v", record)
	}

	snapshot := l.Snapshot()

	if snapshot.State != lobby.StateReady {
		t.Fatalf("expected round reset to StateReady, got %q", snapshot.State)
	}

	if snapshot.Winner != nil {
		t.Fatal("expected winner to be cleared for the new round")
	}

	if snapshot.RoundNumber != 2 {
		t.Fatalf("expected round number 2, got %d", snapshot.RoundNumber)
	}

	if len(snapshot.Scores) != 1 || snapshot.Scores[0].PlayerID != "p1" || snapshot.Scores[0].Points != lobby.DefaultPointsPerRound {
		t.Fatalf("unexpected scores: %+v", snapshot.Scores)
	}

	if len(snapshot.History) != 1 || snapshot.History[0].WinnerID != "p1" {
		t.Fatalf("unexpected history: %+v", snapshot.History)
	}

	// The same roster should still be able to play a second round
	// without rejoining.
	if err := l.OpenBuzz(); err != nil {
		t.Fatal(err)
	}

	if _, err := l.Buzz("p2"); err != nil {
		t.Fatal(err)
	}

	if _, err := l.NextRound(); err != nil {
		t.Fatal(err)
	}

	snapshot = l.Snapshot()

	if len(snapshot.Scores) != 2 {
		t.Fatalf("expected both players to have a score after 2 rounds, got %+v", snapshot.Scores)
	}

	for _, s := range snapshot.Scores {
		if s.Points != lobby.DefaultPointsPerRound {
			t.Fatalf("expected each winner to have exactly %d point, got %+v", lobby.DefaultPointsPerRound, s)
		}
	}

	if len(snapshot.History) != 2 {
		t.Fatalf("expected 2 rounds in history, got %d", len(snapshot.History))
	}
}

func TestNextRoundRejectedBeforeRoundFinishes(t *testing.T) {
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	if _, err := l.NextRound(); !errors.Is(err, lobby.ErrRoundNotFinished) {
		t.Fatalf("expected ErrRoundNotFinished, got %v", err)
	}

	if err := l.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
	}

	if err := l.SetReady(); err != nil {
		t.Fatal(err)
	}

	if _, err := l.NextRound(); !errors.Is(err, lobby.ErrRoundNotFinished) {
		t.Fatalf("expected ErrRoundNotFinished while ready (not locked), got %v", err)
	}
}

func TestUpdateSettings(t *testing.T) {
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	// Should be able to update while waiting.
	points := 3
	countdown := 10

	if err := l.UpdateSettings(lobby.SettingsUpdate{
		PointsPerRound:   &points,
		CountdownSeconds: &countdown,
	}); err != nil {
		t.Fatal(err)
	}

	snap := l.Snapshot()

	if snap.Settings.PointsPerRound != 3 {
		t.Fatalf("expected 3 points, got %d", snap.Settings.PointsPerRound)
	}

	if snap.Settings.CountdownSeconds != 10 {
		t.Fatalf("expected 10s countdown, got %d", snap.Settings.CountdownSeconds)
	}
}

func TestUpdateSettingsRejectedDuringRound(t *testing.T) {
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	if err := l.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
	}

	if err := l.SetReady(); err != nil {
		t.Fatal(err)
	}

	if err := l.OpenBuzz(); err != nil {
		t.Fatal(err)
	}

	points := 5

	if err := l.UpdateSettings(lobby.SettingsUpdate{PointsPerRound: &points}); !errors.Is(err, lobby.ErrSettingsLocked) {
		t.Fatalf("expected ErrSettingsLocked while open, got %v", err)
	}
}

func TestUpdateSettingsValidation(t *testing.T) {
	l := lobby.New("abc", "test", "host", true, lobby.DefaultSettings())

	tooHigh := lobby.MaxPointsPerRound + 1

	if err := l.UpdateSettings(lobby.SettingsUpdate{PointsPerRound: &tooHigh}); !errors.Is(err, lobby.ErrInvalidSettings) {
		t.Fatalf("expected ErrInvalidSettings for points > max, got %v", err)
	}

	zero := 0

	if err := l.UpdateSettings(lobby.SettingsUpdate{PointsPerRound: &zero}); err != nil {
		t.Fatalf("zero points should be valid (no-score mode), got %v", err)
	}

	tooLongCountdown := lobby.MaxCountdownSeconds + 1

	if err := l.UpdateSettings(lobby.SettingsUpdate{CountdownSeconds: &tooLongCountdown}); !errors.Is(err, lobby.ErrInvalidSettings) {
		t.Fatalf("expected ErrInvalidSettings for countdown > max, got %v", err)
	}

	zeroCountdown := 0

	if err := l.UpdateSettings(lobby.SettingsUpdate{CountdownSeconds: &zeroCountdown}); err != nil {
		t.Fatalf("zero countdown should be valid (instant open mode), got %v", err)
	}
}

func TestCustomPointsAwardedOnNextRound(t *testing.T) {
	settings := lobby.LobbySettings{PointsPerRound: 5, CountdownSeconds: 0}
	l := lobby.New("abc", "test", "host", true, settings)

	if err := l.AddPlayer(&lobby.Player{ID: "p1"}); err != nil {
		t.Fatal(err)
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

	record, err := l.NextRound()

	if err != nil {
		t.Fatal(err)
	}

	if record.Points != 5 {
		t.Fatalf("expected 5 points per round, got %d", record.Points)
	}

	snap := l.Snapshot()

	if len(snap.Scores) != 1 || snap.Scores[0].Points != 5 {
		t.Fatalf("expected score of 5, got %+v", snap.Scores)
	}
}
