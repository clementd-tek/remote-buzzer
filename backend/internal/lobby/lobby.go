package lobby

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type State string

const (
	StateWaiting State = "waiting"
	StateReady   State = "ready"
	StateOpen    State = "open"
	StateLocked  State = "locked"

	// MaxNameLength bounds lobby names and player names coming from
	// clients.
	MaxNameLength = 60
)

var (
	ErrLobbyClosed     = errors.New("lobby is not open")
	ErrAlreadyBuzzed   = errors.New("someone already buzzed")
	ErrRoundInProgress = errors.New("cannot join while a round is in progress")
	ErrPlayerNotFound  = errors.New("player is not in this lobby")
	ErrInvalidName     = errors.New("name must not be empty")
	ErrInvalidID       = errors.New("id must not be empty")
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
	ID        string
	Name      string
	Public    bool
	mu        sync.RWMutex
	State     State
	HostID    string
	Players   map[string]*Player
	Winner    *BuzzResult
	UpdatedAt time.Time
}

type LobbySnapshot struct {
	ID          string
	Name        string
	Public      bool
	State       State
	HostID      string
	PlayerCount int
	Players     []PlayerInfo
	Winner      *BuzzResult
	UpdatedAt   time.Time
}

// PlayerInfo is the public-facing view of a Player: enough to render a
// roster without leaking anything internal.
type PlayerInfo struct {
	ID   string
	Name string
}

func New(id string, name string, hostID string, public bool) *Lobby {
	return &Lobby{
		ID:        id,
		Name:      name,
		Public:    public,
		State:     StateWaiting,
		HostID:    hostID,
		Players:   make(map[string]*Player),
		UpdatedAt: time.Now(),
	}
}

// Restore rebuilds a Lobby from a previously cached snapshot (e.g. after a
// backend restart). Session-only data such as the exact player list is not
// preserved: it is intentionally ephemeral, since a live game round is tied
// to the websocket connections held by a single running process anyway.
// The lobby directory (id, name, state, host, counts) does survive.
func Restore(snapshot LobbySnapshot) *Lobby {
	return &Lobby{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Public:    snapshot.Public,
		State:     snapshot.State,
		HostID:    snapshot.HostID,
		Players:   make(map[string]*Player),
		Winner:    snapshot.Winner,
		UpdatedAt: snapshot.UpdatedAt,
	}
}

func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}

	if len(name) > MaxNameLength {
		return ErrInvalidName
	}

	return nil
}

func ValidateID(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}

	return nil
}

func (l *Lobby) AddPlayer(player *Player) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Once a round has started, late joiners would either miss the buzz
	// window entirely or unfairly join after the fact; block it instead.
	if l.State == StateOpen || l.State == StateLocked {
		return ErrRoundInProgress
	}

	if player.JoinedAt.IsZero() {
		player.JoinedAt = time.Now()
	}

	l.Players[player.ID] = player
	l.UpdatedAt = time.Now()

	return nil
}

func (l *Lobby) SetReady() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateWaiting {

		return ErrLobbyClosed
	}

	l.State = StateReady
	l.UpdatedAt = time.Now()

	return nil
}

func (l *Lobby) OpenBuzz() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateReady {

		return ErrLobbyClosed
	}

	l.State = StateOpen
	l.UpdatedAt = time.Now()

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

	if _, ok := l.Players[playerID]; !ok {
		return nil, ErrPlayerNotFound
	}

	result := &BuzzResult{
		PlayerID: playerID,
		Time:     time.Now(),
	}

	l.Winner = result
	l.State = StateLocked
	l.UpdatedAt = time.Now()

	return result, nil
}

// IsStale reports whether the lobby has had no activity for longer than
// ttl, making it a candidate for cleanup.
func (l *Lobby) IsStale(ttl time.Duration) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return time.Since(l.UpdatedAt) > ttl
}

func (l *Lobby) Snapshot() LobbySnapshot {
	l.mu.RLock()
	defer l.mu.RUnlock()

	players := make([]PlayerInfo, 0, len(l.Players))

	for _, p := range l.Players {
		players = append(players, PlayerInfo{ID: p.ID, Name: p.Name})
	}

	sort.Slice(players, func(i, j int) bool {
		return l.Players[players[i].ID].JoinedAt.Before(l.Players[players[j].ID].JoinedAt)
	})

	return LobbySnapshot{
		ID:          l.ID,
		Name:        l.Name,
		Public:      l.Public,
		State:       l.State,
		HostID:      l.HostID,
		PlayerCount: len(l.Players),
		Players:     players,
		Winner:      l.Winner,
		UpdatedAt:   l.UpdatedAt,
	}
}
