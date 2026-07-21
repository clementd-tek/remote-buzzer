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
	StateWaiting   State = "waiting"
	StateReady     State = "ready"
	StateCountdown State = "countdown"
	StateOpen      State = "open"
	StateLocked    State = "locked"

	// MaxNameLength bounds lobby names and player names coming from
	// clients.
	MaxNameLength = 60

	// DefaultPointsPerRound and DefaultCountdownSeconds are used when a
	// lobby is created without explicit settings.
	DefaultPointsPerRound   = 1
	DefaultCountdownSeconds = 3

	// MaxPointsPerRound bounds how many points a single round can be
	// worth. Zero is allowed (play with no scoring at all).
	MaxPointsPerRound = 20

	// MaxCountdownSeconds bounds how long a host can make the pre-buzz
	// countdown. Zero is allowed (open instantly, no countdown).
	MaxCountdownSeconds = 30
)

var (
	ErrLobbyClosed      = errors.New("lobby is not open")
	ErrAlreadyBuzzed    = errors.New("someone already buzzed")
	ErrRoundInProgress  = errors.New("cannot join while a round is in progress")
	ErrPlayerNotFound   = errors.New("player is not in this lobby")
	ErrInvalidName      = errors.New("name must not be empty")
	ErrInvalidID        = errors.New("id must not be empty")
	ErrRoundNotFinished = errors.New("current round has not finished yet")
	ErrInvalidSettings  = errors.New("invalid lobby settings")
	ErrSettingsLocked   = errors.New("settings can't change once a round is underway")
)

// LobbySettings are the host-configurable rules for a lobby: how many
// points a round win is worth, and how long the pre-buzz countdown runs.
type LobbySettings struct {
	PointsPerRound   int
	CountdownSeconds int
}

// DefaultSettings returns the settings used when a lobby is created
// without explicit ones.
func DefaultSettings() LobbySettings {
	return LobbySettings{
		PointsPerRound:   DefaultPointsPerRound,
		CountdownSeconds: DefaultCountdownSeconds,
	}
}

// ValidateSettings checks that settings are within the bounds the rest
// of the package assumes.
func ValidateSettings(s LobbySettings) error {
	if s.PointsPerRound < 0 || s.PointsPerRound > MaxPointsPerRound {
		return ErrInvalidSettings
	}

	if s.CountdownSeconds < 0 || s.CountdownSeconds > MaxCountdownSeconds {
		return ErrInvalidSettings
	}

	return nil
}

// SettingsUpdate patches a subset of LobbySettings; a nil field means
// "leave this one as it is", which lets a host change just the points or
// just the countdown without needing to resend both.
type SettingsUpdate struct {
	PointsPerRound   *int
	CountdownSeconds *int
}

type Player struct {
	ID       string
	Name     string
	JoinedAt time.Time
}

type BuzzResult struct {
	PlayerID string
	Time     time.Time
}

// RoundRecord captures the outcome of one finished round, kept so the
// lobby can show a short recap ("who won each round") across a multi-
// round session.
type RoundRecord struct {
	Round      int
	WinnerID   string
	WinnerName string
	Points     int
	ClosedAt   time.Time
}

type Lobby struct {
	ID              string
	Name            string
	Public          bool
	mu              sync.RWMutex
	State           State
	HostID          string
	Players         map[string]*Player
	Winner          *BuzzResult
	UpdatedAt       time.Time
	RoundNumber     int
	CountdownEndsAt *time.Time
	Scores          map[string]int
	History         []RoundRecord
	Settings        LobbySettings
}

type LobbySnapshot struct {
	ID              string
	Name            string
	Public          bool
	State           State
	HostID          string
	PlayerCount     int
	Players         []PlayerInfo
	Winner          *BuzzResult
	UpdatedAt       time.Time
	RoundNumber     int
	CountdownEndsAt *time.Time
	Scores          []ScoreEntry
	History         []RoundRecord
	Settings        LobbySettings
}

// PlayerInfo is the public-facing view of a Player: enough to render a
// roster without leaking anything internal.
type PlayerInfo struct {
	ID   string
	Name string
}

// ScoreEntry is one player's cumulative score, already resolved against
// the player's current display name so callers don't need to join
// against the player list themselves.
type ScoreEntry struct {
	PlayerID string
	Name     string
	Points   int
}

// New creates a lobby with the given settings. Use DefaultSettings() if
// the caller doesn't have anything specific in mind.
func New(id string, name string, hostID string, public bool, settings LobbySettings) *Lobby {
	return &Lobby{
		ID:          id,
		Name:        name,
		Public:      public,
		State:       StateWaiting,
		HostID:      hostID,
		Players:     make(map[string]*Player),
		UpdatedAt:   time.Now(),
		RoundNumber: 1,
		Scores:      make(map[string]int),
		History:     make([]RoundRecord, 0),
		Settings:    settings,
	}
}

// Restore rebuilds a Lobby from a previously cached snapshot (e.g. after a
// backend restart). Session-only data such as the exact player list is not
// preserved: it is intentionally ephemeral, since a live game round is tied
// to the websocket connections held by a single running process anyway.
// The lobby directory (id, name, state, host, counts) does survive, and so
// does game progress (round number, scores, history, settings) since
// that's real progress worth keeping rather than session plumbing.
func Restore(snapshot LobbySnapshot) *Lobby {
	scores := make(map[string]int, len(snapshot.Scores))

	for _, entry := range snapshot.Scores {
		scores[entry.PlayerID] = entry.Points
	}

	roundNumber := snapshot.RoundNumber
	if roundNumber == 0 {
		roundNumber = 1
	}

	history := snapshot.History
	if history == nil {
		history = make([]RoundRecord, 0)
	}

	settings := snapshot.Settings
	if settings == (LobbySettings{}) {
		settings = DefaultSettings()
	}

	return &Lobby{
		ID:          snapshot.ID,
		Name:        snapshot.Name,
		Public:      snapshot.Public,
		State:       snapshot.State,
		HostID:      snapshot.HostID,
		Players:     make(map[string]*Player),
		Winner:      snapshot.Winner,
		UpdatedAt:   snapshot.UpdatedAt,
		RoundNumber: roundNumber,
		Scores:      scores,
		History:     history,
		Settings:    settings,
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
	if l.State == StateCountdown || l.State == StateOpen || l.State == StateLocked {
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

// UpdateSettings applies a partial settings update. Only allowed while
// nothing is actively in flight (waiting or between rounds) so the
// rules can't change out from under a round that's already running.
func (l *Lobby) UpdateSettings(update SettingsUpdate) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateWaiting && l.State != StateReady {
		return ErrSettingsLocked
	}

	next := l.Settings

	if update.PointsPerRound != nil {
		next.PointsPerRound = *update.PointsPerRound
	}

	if update.CountdownSeconds != nil {
		next.CountdownSeconds = *update.CountdownSeconds
	}

	if err := ValidateSettings(next); err != nil {
		return err
	}

	l.Settings = next
	l.UpdatedAt = time.Now()

	return nil
}

// StartCountdown begins the pre-buzz countdown, ending at endsAt. Clients
// render the countdown themselves from this absolute timestamp (rather
// than a relative "starting now" signal) so it stays in sync regardless
// of network latency or when a given client's message actually arrives.
func (l *Lobby) StartCountdown(endsAt time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateReady {
		return ErrLobbyClosed
	}

	l.State = StateCountdown
	l.CountdownEndsAt = &endsAt
	l.UpdatedAt = time.Now()

	return nil
}

// OpenBuzz opens the buzzer for a round, allowing players to buzz in. It
// accepts either a direct ready->open transition (instant open, no
// countdown) or countdown->open (the normal path once a countdown
// finishes).
func (l *Lobby) OpenBuzz() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateReady && l.State != StateCountdown {
		return ErrLobbyClosed
	}

	l.State = StateOpen
	l.CountdownEndsAt = nil
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

// NextRound closes out the current (finished) round — awarding points to
// its winner (per the lobby's configured PointsPerRound) and recording it
// in the round history — then resets the lobby to StateReady with the
// same roster so the host can immediately start another round. Returns
// the record of the round that just closed (nil if nobody buzzed in that
// round, e.g. it was skipped).
func (l *Lobby) NextRound() (*RoundRecord, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateLocked {
		return nil, ErrRoundNotFinished
	}

	var record *RoundRecord

	if l.Winner != nil {
		if l.Scores == nil {
			l.Scores = make(map[string]int)
		}

		points := l.Settings.PointsPerRound
		l.Scores[l.Winner.PlayerID] += points

		winnerName := ""
		if p, ok := l.Players[l.Winner.PlayerID]; ok {
			winnerName = p.Name
		}

		record = &RoundRecord{
			Round:      l.RoundNumber,
			WinnerID:   l.Winner.PlayerID,
			WinnerName: winnerName,
			Points:     points,
			ClosedAt:   time.Now(),
		}

		l.History = append(l.History, *record)
	}

	l.RoundNumber++
	l.State = StateReady
	l.Winner = nil
	l.CountdownEndsAt = nil
	l.UpdatedAt = time.Now()

	return record, nil
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

	scores := make([]ScoreEntry, 0, len(l.Scores))

	for playerID, points := range l.Scores {
		name := ""
		if p, ok := l.Players[playerID]; ok {
			name = p.Name
		}

		scores = append(scores, ScoreEntry{PlayerID: playerID, Name: name, Points: points})
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Points != scores[j].Points {
			return scores[i].Points > scores[j].Points
		}

		// Stable tie-break so the ordering doesn't jitter between
		// snapshots when points are equal.
		return scores[i].PlayerID < scores[j].PlayerID
	})

	history := l.History
	if history == nil {
		history = make([]RoundRecord, 0)
	}

	var countdownEndsAt *time.Time
	if l.CountdownEndsAt != nil {
		t := *l.CountdownEndsAt
		countdownEndsAt = &t
	}

	return LobbySnapshot{
		ID:              l.ID,
		Name:            l.Name,
		Public:          l.Public,
		State:           l.State,
		HostID:          l.HostID,
		PlayerCount:     len(l.Players),
		Players:         players,
		Winner:          l.Winner,
		UpdatedAt:       l.UpdatedAt,
		RoundNumber:     l.RoundNumber,
		CountdownEndsAt: countdownEndsAt,
		Scores:          scores,
		History:         history,
		Settings:        l.Settings,
	}
}
