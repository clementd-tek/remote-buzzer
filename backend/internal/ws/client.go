package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Inbound is a message sent by a connected client (a player or the host).
type Inbound struct {
	Type     string `json:"type"`
	PlayerID string `json:"playerId,omitempty"`

	// Settings fields — only read for type == "settings".
	PointsPerRound   *int `json:"pointsPerRound,omitempty"`
	CountdownSeconds *int `json:"countdownSeconds,omitempty"`
}

// Client represents a single websocket connection belonging to a player
// (or the host) inside a lobby.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	lobbyID  string
	playerID string
	logger   *slog.Logger

	onMessage func(c *Client, msg Inbound)
}

func newClient(hub *Hub, conn *websocket.Conn, lobbyID string, playerID string, logger *slog.Logger, onMessage func(c *Client, msg Inbound)) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 16),
		lobbyID:   lobbyID,
		playerID:  playerID,
		logger:    logger,
		onMessage: onMessage,
	}
}

// PlayerID returns the id the client identified itself with when it
// connected (?playerId=... query param).
func (c *Client) PlayerID() string {
	return c.playerID
}

// Send marshals msg to JSON and queues it for delivery to this client
// only (as opposed to Hub.Broadcast, which reaches the whole room).
func (c *Client) Send(msg any) {
	payload, err := json.Marshal(msg)

	if err != nil {
		if c.logger != nil {
			c.logger.Error("ws: failed to marshal message", "error", err)
		}
		return
	}

	select {
	case c.send <- payload:
	default:
		// buffer full, client is too slow: drop the message rather than
		// block the caller.
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c.lobbyID, c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()

		if err != nil {
			return
		}

		var msg Inbound

		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		if msg.PlayerID == "" {
			msg.PlayerID = c.playerID
		}

		if c.onMessage != nil {
			c.onMessage(c, msg)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case payload, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
