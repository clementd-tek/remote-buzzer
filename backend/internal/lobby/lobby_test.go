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

	l.AddPlayer(
		&lobby.Player{
			ID: "p1",
		},
	)

	// manually move state
	l.OpenBuzz()

	result, err := l.Buzz("p1")

	if err != nil {
		t.Fatal(err)
	}

	if result.PlayerID != "p1" {
		t.Fatal("wrong winner")
	}
}
