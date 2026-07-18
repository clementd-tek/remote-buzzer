package lobby

import (
	"errors"
	"sync"
	"time"
)

type State string

const (
	StateWaiting State = "waiting"
	StateReady   State = "ready"
	StateOpen    State = "open"
	StateLocked  State = "locked"
)

var (
	ErrLobbyClosed   = errors.New("lobby is not open")
	ErrAlreadyBuzzed = errors.New("someone already buzzed")
)

type Player struct {
	ID       string
	Name     string
	JoinedAt time.Time
}

type BuzzResult struct {
	PlayerID string
	Time     time.Time
}

type Lobby struct {
	ID      string
	Name    string
	Public  bool
	mu      sync.RWMutex
	State   State
	HostID  string
	Players map[string]*Player
	Winner  *BuzzResult
}

type LobbySnapshot struct {
	ID          string
	Name        string
	Public      bool
	State       State
	HostID      string
	PlayerCount int
	Winner      *BuzzResult
}

func New(id string, name string, hostID string, public bool) *Lobby {
	return &Lobby{
		ID:      id,
		Name:    name,
		Public:  public,
		State:   StateWaiting,
		HostID:  hostID,
		Players: make(map[string]*Player),
	}
}

func (l *Lobby) AddPlayer(player *Player) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Players[player.ID] = player

	return nil
}

func (l *Lobby) SetReady() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateWaiting {

		return ErrLobbyClosed
	}

	l.State = StateReady

	return nil
}

func (l *Lobby) OpenBuzz() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateReady {

		return ErrLobbyClosed
	}

	l.State = StateOpen

	return nil
}

func (l *Lobby) Buzz(playerID string) (*BuzzResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateOpen {
		return nil, ErrLobbyClosed
	}

	if l.Winner != nil {
		return nil, ErrAlreadyBuzzed
	}

	result := &BuzzResult{
		PlayerID: playerID,
		Time:     time.Now(),
	}

	l.Winner = result
	l.State = StateLocked

	return result, nil
}

func (l *Lobby) Snapshot() LobbySnapshot {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return LobbySnapshot{
		ID:          l.ID,
		Name:        l.Name,
		Public:      l.Public,
		State:       l.State,
		HostID:      l.HostID,
		PlayerCount: len(l.Players),
		Winner:      l.Winner,
	}
}
