package handlers

import (
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api/dto"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/clementd-tek/remote-buzzer/backend/internal/ws"

	"github.com/go-chi/chi/v5"
)

// wsOutbound is every message the server can push down a lobby's
// websocket connections.
type wsOutbound struct {
	Type  string             `json:"type"`
	Lobby *dto.LobbyResponse `json:"lobby,omitempty"`
	Error string             `json:"error,omitempty"`
}

type LobbyWSHandler struct {
	service *lobby.Service
	hub     *ws.Hub
}

func NewLobbyWSHandler(service *lobby.Service, hub *ws.Hub) *LobbyWSHandler {
	return &LobbyWSHandler{
		service: service,
		hub:     hub,
	}
}

// Serve upgrades GET /api/lobbies/{id}/ws?playerId=... to a websocket
// connection. Every player (and the host) connects here to receive live
// lobby updates and to send buzzer actions.
func (h *LobbyWSHandler) Serve(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	playerID := r.URL.Query().Get("playerId")

	l, err := h.service.Get(lobbyID)

	if err != nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}

	client, err := h.hub.Serve(
		w,
		r,
		lobbyID,
		playerID,
		func(c *ws.Client, msg ws.Inbound) {
			h.handleMessage(lobbyID, l, c, msg)
		},
	)

	if err != nil {
		return
	}

	// Greet the newly connected client with the current lobby state, and
	// let everyone else in the room know the player count may have
	// changed.
	client.Send(wsOutbound{
		Type:  "lobby_update",
		Lobby: lobbyResponse(l),
	})

	h.hub.Broadcast(lobbyID, wsOutbound{
		Type:  "lobby_update",
		Lobby: lobbyResponse(l),
	})
}

func (h *LobbyWSHandler) handleMessage(lobbyID string, l *lobby.Lobby, c *ws.Client, msg ws.Inbound) {
	switch msg.Type {
	case "ready":
		if msg.PlayerID != l.HostID {
			c.Send(wsOutbound{Type: "error", Error: "only the host can ready the lobby"})
			return
		}

		if err := l.SetReady(); err != nil {
			c.Send(wsOutbound{Type: "error", Error: err.Error()})
			return
		}

	case "open":
		if msg.PlayerID != l.HostID {
			c.Send(wsOutbound{Type: "error", Error: "only the host can open the buzzer"})
			return
		}

		if err := l.OpenBuzz(); err != nil {
			c.Send(wsOutbound{Type: "error", Error: err.Error()})
			return
		}

	case "buzz":
		if _, err := l.Buzz(msg.PlayerID); err != nil {
			c.Send(wsOutbound{Type: "error", Error: err.Error()})
			return
		}

	default:
		c.Send(wsOutbound{Type: "error", Error: "unknown message type"})
		return
	}

	h.hub.Broadcast(lobbyID, wsOutbound{
		Type:  "lobby_update",
		Lobby: lobbyResponse(l),
	})
}

func lobbyResponse(l *lobby.Lobby) *dto.LobbyResponse {
	response := dto.FromLobby(l.Snapshot())
	return &response
}
