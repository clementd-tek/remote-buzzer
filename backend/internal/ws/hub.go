package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/clementd-tek/remote-buzzer/backend/internal/originpolicy"
	"github.com/gorilla/websocket"
)

// room groups every client currently connected to a given lobby.
type room struct {
	clients map[*Client]bool
}

// Hub keeps track of all per-lobby rooms and lets the rest of the
// application broadcast messages to every client connected to a lobby.
type Hub struct {
	mu             sync.RWMutex
	rooms          map[string]*room
	logger         *slog.Logger
	allowedOrigins map[string]bool
	upgrader       websocket.Upgrader
}

// NewHub creates a Hub. allowedOrigins lists the origins permitted to open
// a websocket connection (e.g. the frontend's URL); a request without an
// Origin header (non-browser clients, same-origin tooling) is always
// allowed since browsers are the only clients that send it.
func NewHub(logger *slog.Logger, allowedOrigins []string) *Hub {
	originSet := make(map[string]bool, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		if origin != "" {
			originSet[origin] = true
		}
	}

	h := &Hub{
		rooms:          make(map[string]*room),
		logger:         logger,
		allowedOrigins: originSet,
	}

	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}

	return h
}

func (h *Hub) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	if origin == "" {
		return true
	}

	if originpolicy.IsLocal(origin) {
		return true
	}

	if len(h.allowedOrigins) == 0 {
		return true
	}

	if h.allowedOrigins[origin] {
		return true
	}

	if h.logger != nil {
		h.logger.Warn(
			"ws: rejected connection from unrecognized origin",
			"origin", origin,
			"allowed", h.allowedOrigins,
		)
	}

	return false
}

// Serve upgrades the HTTP connection to a websocket, registers the
// resulting client under lobbyID and starts its read/write pumps.
// onMessage is invoked (from the client's read loop) for every message
// sent by that client.
func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, lobbyID string, playerID string, onMessage func(c *Client, msg Inbound)) (*Client, error) {
	conn, err := h.upgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, err
	}

	client := newClient(h, conn, lobbyID, playerID, h.logger, onMessage)

	h.register(lobbyID, client)

	go client.writePump()
	go client.readPump()

	return client, nil
}

func (h *Hub) register(lobbyID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	r, ok := h.rooms[lobbyID]

	if !ok {
		r = &room{clients: make(map[*Client]bool)}
		h.rooms[lobbyID] = r
	}

	r.clients[c] = true
}

func (h *Hub) unregister(lobbyID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	r, ok := h.rooms[lobbyID]

	if !ok {
		return
	}

	if _, ok := r.clients[c]; ok {
		delete(r.clients, c)
		close(c.send)
	}

	if len(r.clients) == 0 {
		delete(h.rooms, lobbyID)
	}
}

// Broadcast sends a JSON-encoded message to every client connected to a
// lobby. Slow or dead clients are dropped instead of blocking everyone
// else.
func (h *Hub) Broadcast(lobbyID string, msg any) {
	payload, err := json.Marshal(msg)

	if err != nil {
		if h.logger != nil {
			h.logger.Error("ws: failed to marshal broadcast", "error", err)
		}
		return
	}

	h.mu.RLock()
	r, ok := h.rooms[lobbyID]

	if !ok {
		h.mu.RUnlock()
		return
	}

	stuck := make([]*Client, 0)

	for c := range r.clients {
		select {
		case c.send <- payload:
		default:
			stuck = append(stuck, c)
		}
	}

	h.mu.RUnlock()

	for _, c := range stuck {
		h.unregister(lobbyID, c)
	}
}
