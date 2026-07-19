package lobby_test

import (
	"errors"
	"testing"

	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
)

func newTestService() *lobby.Service {
	manager := lobby.NewManager(nil, discardLogger())
	return lobby.NewService(manager)
}

func TestServiceCreateValidatesInput(t *testing.T) {
	service := newTestService()

	if _, err := service.Create(lobby.CreateLobbyRequest{Name: "", HostID: "host"}); !errors.Is(err, lobby.ErrInvalidName) {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}

	if _, err := service.Create(lobby.CreateLobbyRequest{Name: "Quiz", HostID: ""}); !errors.Is(err, lobby.ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestServiceListPublicFiltersPrivateLobbies(t *testing.T) {
	service := newTestService()

	if _, err := service.Create(lobby.CreateLobbyRequest{Name: "Public quiz", HostID: "host1", Public: true}); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Create(lobby.CreateLobbyRequest{Name: "Private quiz", HostID: "host2", Public: false}); err != nil {
		t.Fatal(err)
	}

	public := service.ListPublic()

	if len(public) != 1 {
		t.Fatalf("expected 1 public lobby, got %d", len(public))
	}

	if public[0].Name != "Public quiz" {
		t.Fatalf("expected the public lobby, got %q", public[0].Name)
	}

	if len(service.List()) != 2 {
		t.Fatal("List() should still return every lobby, public and private")
	}
}

func TestServiceRoundFlow(t *testing.T) {
	service := newTestService()

	l, err := service.Create(lobby.CreateLobbyRequest{Name: "Quiz", HostID: "host", Public: true})

	if err != nil {
		t.Fatal(err)
	}

	if err := service.Join(l.ID, &lobby.Player{ID: "p1", Name: "Alice"}); err != nil {
		t.Fatal(err)
	}

	if _, err := service.SetReady(l.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := service.OpenBuzz(l.ID); err != nil {
		t.Fatal(err)
	}

	updated, err := service.Buzz(l.ID, "p1")

	if err != nil {
		t.Fatal(err)
	}

	snapshot := updated.Snapshot()

	if snapshot.Winner == nil || snapshot.Winner.PlayerID != "p1" {
		t.Fatal("expected p1 to be recorded as the winner")
	}
}

func TestServiceJoinRejectsUnknownLobby(t *testing.T) {
	service := newTestService()

	err := service.Join("does-not-exist", &lobby.Player{ID: "p1", Name: "Alice"})

	if !errors.Is(err, lobby.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
