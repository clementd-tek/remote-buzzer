package dto

import (
	"time"

	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
)

type LobbyResponse struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Public          bool             `json:"public"`
	State           string           `json:"state"`
	HostID          string           `json:"hostId"`
	PlayerCount     int              `json:"playerCount"`
	Players         []PlayerResponse `json:"players"`
	Winner          *WinnerResponse  `json:"winner,omitempty"`
	RoundNumber     int              `json:"roundNumber"`
	CountdownEndsAt *time.Time       `json:"countdownEndsAt,omitempty"`
	Scores          []ScoreResponse  `json:"scores"`
	History         []RoundResponse  `json:"history"`
}

type WinnerResponse struct {
	PlayerID string    `json:"playerId"`
	Time     time.Time `json:"time"`
}

// ScoreResponse is one player's cumulative points, sorted descending by
// the backend before being sent (see lobby.Lobby.Snapshot).
type ScoreResponse struct {
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Points   int    `json:"points"`
}

// RoundResponse is the outcome of one finished round, used to render a
// short "previous rounds" recap.
type RoundResponse struct {
	Round      int       `json:"round"`
	WinnerID   string    `json:"winnerId"`
	WinnerName string    `json:"winnerName"`
	Points     int       `json:"points"`
	ClosedAt   time.Time `json:"closedAt"`
}

func FromLobby(l lobby.LobbySnapshot) LobbyResponse {
	players := make([]PlayerResponse, 0, len(l.Players))

	for _, p := range l.Players {
		players = append(players, PlayerResponse{ID: p.ID, Name: p.Name})
	}

	scores := make([]ScoreResponse, 0, len(l.Scores))

	for _, s := range l.Scores {
		scores = append(scores, ScoreResponse{PlayerID: s.PlayerID, Name: s.Name, Points: s.Points})
	}

	history := make([]RoundResponse, 0, len(l.History))

	for _, r := range l.History {
		history = append(history, RoundResponse{
			Round:      r.Round,
			WinnerID:   r.WinnerID,
			WinnerName: r.WinnerName,
			Points:     r.Points,
			ClosedAt:   r.ClosedAt,
		})
	}

	response := LobbyResponse{
		ID:          l.ID,
		Name:        l.Name,
		Public:      l.Public,
		State:       string(l.State),
		HostID:      l.HostID,
		PlayerCount: l.PlayerCount,
		Players:     players,
		RoundNumber: l.RoundNumber,
		Scores:      scores,
		History:     history,
	}

	if l.Winner != nil {
		response.Winner = &WinnerResponse{
			PlayerID: l.Winner.PlayerID,
			Time:     l.Winner.Time,
		}
	}

	if l.CountdownEndsAt != nil {
		t := *l.CountdownEndsAt
		response.CountdownEndsAt = &t
	}

	return response
}
