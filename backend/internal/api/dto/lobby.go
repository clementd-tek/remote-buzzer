package dto

import (
	"time"

	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
)

type LobbyResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Public      bool            `json:"public"`
	State       string          `json:"state"`
	HostID      string          `json:"hostId"`
	PlayerCount int             `json:"playerCount"`
	Winner      *WinnerResponse `json:"winner,omitempty"`
}

type WinnerResponse struct {
	PlayerID string    `json:"playerId"`
	Time     time.Time `json:"time"`
}

func FromLobby(l lobby.LobbySnapshot) LobbyResponse {
	response := LobbyResponse{
		ID:          l.ID,
		Name:        l.Name,
		Public:      l.Public,
		State:       string(l.State),
		HostID:      l.HostID,
		PlayerCount: l.PlayerCount,
	}

	if l.Winner != nil {
		response.Winner = &WinnerResponse{
			PlayerID: l.Winner.PlayerID,
			Time:     l.Winner.Time,
		}
	}

	return response
}
