package handlers

import (
	"net/http"
	"time"

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
// connection.
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

	h.broadcastLobby(lobbyID, l)
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

		h.broadcastLobby(lobbyID, l)

	case "open":
		if msg.PlayerID != l.HostID {
			c.Send(wsOutbound{Type: "error", Error: "only the host can open the buzzer"})
			return
		}

		// The countdown duration comes from the lobby's own settings,
		// not from the client message — this way the duration is always
		// what the host configured via "settings", not whatever a
		// client could send in the "open" payload.
		snap := l.Snapshot()
		h.handleOpen(lobbyID, l, c, snap.Settings.CountdownSeconds)

	case "buzz":
		if _, err := l.Buzz(msg.PlayerID); err != nil {
			c.Send(wsOutbound{Type: "error", Error: err.Error()})
			return
		}

		h.broadcastLobby(lobbyID, l)

	case "next_round":
		if msg.PlayerID != l.HostID {
			c.Send(wsOutbound{Type: "error", Error: "only the host can start the next round"})
			return
		}

		if _, err := l.NextRound(); err != nil {
			c.Send(wsOutbound{Type: "error", Error: err.Error()})
			return
		}

		h.broadcastLobby(lobbyID, l)

	case "settings":
		if msg.PlayerID != l.HostID {
			c.Send(wsOutbound{Type: "error", Error: "only the host can change settings"})
			return
		}

		if err := l.UpdateSettings(lobby.SettingsUpdate{
			PointsPerRound:   msg.PointsPerRound,
			CountdownSeconds: msg.CountdownSeconds,
		}); err != nil {
			c.Send(wsOutbound{Type: "error", Error: err.Error()})
			return
		}

		h.broadcastLobby(lobbyID, l)

	default:
		c.Send(wsOutbound{Type: "error", Error: "unknown message type"})
	}
}

// handleOpen either opens the buzzer immediately (seconds == 0) or
// schedules it to open after a countdown, broadcasting the countdown
// state right away so every client can render the same "3, 2, 1" using a
// shared server-clock end time.
func (h *LobbyWSHandler) handleOpen(lobbyID string, l *lobby.Lobby, c *ws.Client, seconds int) {
	if seconds <= 0 {
		if err := l.OpenBuzz(); err != nil {
			c.Send(wsOutbound{Type: "error", Error: err.Error()})
			return
		}

		h.broadcastLobby(lobbyID, l)
		return
	}

	if seconds > lobby.MaxCountdownSeconds {
		seconds = lobby.MaxCountdownSeconds
	}

	duration := time.Duration(seconds) * time.Second
	endsAt := time.Now().Add(duration)

	if err := l.StartCountdown(endsAt); err != nil {
		c.Send(wsOutbound{Type: "error", Error: err.Error()})
		return
	}

	h.broadcastLobby(lobbyID, l)

	time.AfterFunc(duration, func() {
		updated, err := h.service.OpenBuzz(lobbyID)

		if err != nil {
			// Lobby may have been deleted (cleanup, restart) in the
			// meantime; nothing to broadcast to.
			return
		}

		h.broadcastLobby(lobbyID, updated)
	})
}

func (h *LobbyWSHandler) broadcastLobby(lobbyID string, l *lobby.Lobby) {
	h.hub.Broadcast(lobbyID, wsOutbound{
		Type:  "lobby_update",
		Lobby: lobbyResponse(l),
	})
}

func lobbyResponse(l *lobby.Lobby) *dto.LobbyResponse {
	response := dto.FromLobby(l.Snapshot())
	return &response
}
