package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api/dto"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"

	"github.com/go-chi/chi/v5"
)

type joinLobbyRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LobbyHandler struct {
	service *lobby.Service
}

func NewLobbyHandler(service *lobby.Service) *LobbyHandler {
	return &LobbyHandler{
		service: service,
	}
}

type createLobbyRequest struct {
	Name   string `json:"name"`
	HostID string `json:"hostId"`
	Public bool   `json:"public"`
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

	result := h.service.Create(
		lobby.CreateLobbyRequest{
			Name:   req.Name,
			HostID: req.HostID,
			Public: req.Public,
		},
	)

	response := dto.FromLobby(
		result.Snapshot(),
	)

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(response)
}

func (h *LobbyHandler) List(w http.ResponseWriter, r *http.Request) {
	lobbies := h.service.List()

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
			http.StatusNotFound,
		)

		return
	}

	response := dto.PlayerResponse{
		ID:   player.ID,
		Name: player.Name,
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
