package lobby_test

import (
	"testing"

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
