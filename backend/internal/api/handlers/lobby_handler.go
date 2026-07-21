package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api/dto"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/clementd-tek/remote-buzzer/backend/internal/ws"

	"github.com/go-chi/chi/v5"
)

type joinLobbyRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LobbyHandler struct {
	service *lobby.Service
	hub     *ws.Hub
}

func NewLobbyHandler(service *lobby.Service, hub *ws.Hub) *LobbyHandler {
	return &LobbyHandler{
		service: service,
		hub:     hub,
	}
}

type createLobbyRequest struct {
	Name     string            `json:"name"`
	HostID   string            `json:"hostId"`
	Public   bool              `json:"public"`
	Settings *lobbySettingsReq `json:"settings,omitempty"`
}

// lobbySettingsReq is the JSON shape for initial/updated settings.
// Using pointers here so clients can omit fields they don't want to
// change (treated as "keep the current/default value").
type lobbySettingsReq struct {
	PointsPerRound   *int `json:"pointsPerRound,omitempty"`
	CountdownSeconds *int `json:"countdownSeconds,omitempty"`
}

// statusForError maps a domain error to the HTTP status code that best
// represents it. Anything not recognised falls back to 500.
func statusForError(err error) int {
	switch {
	case errors.Is(err, lobby.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, lobby.ErrInvalidName),
		errors.Is(err, lobby.ErrInvalidID),
		errors.Is(err, lobby.ErrRoundInProgress),
		errors.Is(err, lobby.ErrInvalidSettings),
		errors.Is(err, lobby.ErrSettingsLocked):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *LobbyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLobbyRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(
			w,
			"invalid body",
			http.StatusBadRequest,
		)
		return
	}

	createReq := lobby.CreateLobbyRequest{
		Name:   req.Name,
		HostID: req.HostID,
		Public: req.Public,
	}

	if req.Settings != nil {
		defaults := lobby.DefaultSettings()

		settings := lobby.LobbySettings{
			PointsPerRound:   defaults.PointsPerRound,
			CountdownSeconds: defaults.CountdownSeconds,
		}

		if req.Settings.PointsPerRound != nil {
			settings.PointsPerRound = *req.Settings.PointsPerRound
		}

		if req.Settings.CountdownSeconds != nil {
			settings.CountdownSeconds = *req.Settings.CountdownSeconds
		}

		createReq.Settings = &settings
	}

	result, err := h.service.Create(createReq)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			statusForError(err),
		)
		return
	}

	response := dto.FromLobby(
		result.Snapshot(),
	)

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(response)
}

// List returns public lobbies only. This backs the homepage list, which
// must never leak private lobbies to unauthenticated visitors.
func (h *LobbyHandler) List(w http.ResponseWriter, r *http.Request) {
	lobbies := h.service.ListPublic()

	responses := make(
		[]dto.LobbyResponse,
		0,
		len(lobbies),
	)

	for _, item := range lobbies {

		responses = append(
			responses,
			dto.FromLobby(
				item.Snapshot(),
			),
		)
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(responses)
}

func (h *LobbyHandler) Join(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(
		r,
		"id",
	)

	var req joinLobbyRequest

	err := json.NewDecoder(r.Body).
		Decode(&req)

	if err != nil {
		http.Error(
			w,
			"invalid body",
			http.StatusBadRequest,
		)

		return
	}

	player := &lobby.Player{
		ID:   req.ID,
		Name: req.Name,
	}

	err = h.service.Join(
		lobbyID,
		player,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			statusForError(err),
		)

		return
	}

	response := dto.PlayerResponse{
		ID:   player.ID,
		Name: player.Name,
	}

	if l, getErr := h.service.Get(lobbyID); getErr == nil {
		h.hub.Broadcast(lobbyID, wsOutbound{
			Type:  "lobby_update",
			Lobby: lobbyResponse(l),
		})
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).
		Encode(response)
}

func (h *LobbyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(
		r,
		"id",
	)

	result, err := h.service.Get(id)

	if err != nil {
		http.Error(
			w,
			"lobby not found",
			http.StatusNotFound,
		)
		return
	}

	response := dto.FromLobby(
		result.Snapshot(),
	)

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(response)
}
