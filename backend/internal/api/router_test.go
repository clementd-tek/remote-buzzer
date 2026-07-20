package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/clementd-tek/remote-buzzer/backend/internal/api"
	"github.com/clementd-tek/remote-buzzer/backend/internal/api/dto"
	"github.com/clementd-tek/remote-buzzer/backend/internal/lobby"
	"github.com/clementd-tek/remote-buzzer/backend/internal/ws"
	"github.com/gorilla/websocket"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestRouter() http.Handler {
	manager := lobby.NewManager(nil, discardLogger())
	service := lobby.NewService(manager)
	hub := ws.NewHub(discardLogger(), nil)

	return api.NewRouter(discardLogger(), service, hub, nil)
}

func doJSON(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader

	if body != nil {
		payload, err := json.Marshal(body)

		if err != nil {
			t.Fatal(err)
		}

		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestCreateLobbyValidation(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name":   "",
		"hostId": "host1",
		"public": true,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty name, got %d", rec.Code)
	}
}

func TestCreateAndGetLobby(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name":   "Quiz night",
		"hostId": "host1",
		"public": true,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var created dto.LobbyResponse

	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	rec = doJSON(t, router, http.MethodGet, "/api/lobbies/"+created.ID, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListOnlyReturnsPublicLobbies(t *testing.T) {
	router := newTestRouter()

	doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name": "Public quiz", "hostId": "host1", "public": true,
	})

	doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name": "Private quiz", "hostId": "host2", "public": false,
	})

	rec := doJSON(t, router, http.MethodGet, "/api/lobbies", nil)

	var lobbies []dto.LobbyResponse

	if err := json.Unmarshal(rec.Body.Bytes(), &lobbies); err != nil {
		t.Fatal(err)
	}

	if len(lobbies) != 1 {
		t.Fatalf("expected 1 public lobby in the listing, got %d", len(lobbies))
	}

	if lobbies[0].Name != "Public quiz" {
		t.Fatalf("expected the public lobby, got %q", lobbies[0].Name)
	}
}

func TestJoinLobbyValidation(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name": "Quiz night", "hostId": "host1", "public": true,
	})

	var created dto.LobbyResponse
	json.Unmarshal(rec.Body.Bytes(), &created)

	// missing name
	rec = doJSON(t, router, http.MethodPost, "/api/lobbies/"+created.ID+"/join", map[string]any{
		"id": "p1", "name": "",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty player name, got %d", rec.Code)
	}

	// unknown lobby
	rec = doJSON(t, router, http.MethodPost, "/api/lobbies/does-not-exist/join", map[string]any{
		"id": "p1", "name": "Alice",
	})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown lobby, got %d", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, http.MethodGet, "/healthz", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// TestWebsocketBuzzFlow exercises the real-time path end-to-end: two
// players connect over websocket, the host readies and opens the round,
// and the first buzz is broadcast to both clients as the winner.
func TestWebsocketBuzzFlow(t *testing.T) {
	router := newTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Create the lobby and join two players over REST first.
	rec := doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name": "Quiz night", "hostId": "host1", "public": true,
	})

	var created dto.LobbyResponse
	json.Unmarshal(rec.Body.Bytes(), &created)

	for _, p := range []struct{ id, name string }{{"p1", "Alice"}, {"p2", "Bob"}} {
		rec = doJSON(t, router, http.MethodPost, "/api/lobbies/"+created.ID+"/join", map[string]any{
			"id": p.id, "name": p.name,
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("join failed for %s: %d %s", p.id, rec.Code, rec.Body.String())
		}
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/lobbies/" + created.ID + "/ws"

	hostConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?playerId=host1", nil)
	if err != nil {
		t.Fatalf("host dial failed: %v", err)
	}
	defer hostConn.Close()

	playerConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?playerId=p1", nil)
	if err != nil {
		t.Fatalf("player dial failed: %v", err)
	}
	defer playerConn.Close()

	// Host readies the lobby and opens the buzzer. waitForState skips
	// over the initial connection snapshots (their exact count varies
	// depending on connection order) and waits for the state we want.
	if err := hostConn.WriteJSON(map[string]string{"type": "ready"}); err != nil {
		t.Fatal(err)
	}
	waitForState(t, hostConn, "ready")

	if err := hostConn.WriteJSON(map[string]string{"type": "open"}); err != nil {
		t.Fatal(err)
	}
	waitForState(t, hostConn, "open")

	// p1 buzzes in.
	if err := playerConn.WriteJSON(map[string]string{"type": "buzz", "playerId": "p1"}); err != nil {
		t.Fatal(err)
	}

	update := waitForState(t, playerConn, "locked")

	if update.Lobby == nil || update.Lobby.Winner == nil || update.Lobby.Winner.PlayerID != "p1" {
		t.Fatalf("expected p1 to win, got %+v", update.Lobby)
	}
}

// TestWebsocketCountdownAndMultipleRounds exercises the full multi-round
// flow: a countdown-gated open, a buzz, next_round awarding a point and
// resetting to ready, then a second round for a different winner.
func TestWebsocketCountdownAndMultipleRounds(t *testing.T) {
	router := newTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	rec := doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name": "Quiz night", "hostId": "host1", "public": true,
	})

	var created dto.LobbyResponse
	json.Unmarshal(rec.Body.Bytes(), &created)

	for _, p := range []struct{ id, name string }{{"p1", "Alice"}, {"p2", "Bob"}} {
		rec = doJSON(t, router, http.MethodPost, "/api/lobbies/"+created.ID+"/join", map[string]any{
			"id": p.id, "name": p.name,
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("join failed for %s: %d %s", p.id, rec.Code, rec.Body.String())
		}
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/lobbies/" + created.ID + "/ws"

	hostConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?playerId=host1", nil)
	if err != nil {
		t.Fatalf("host dial failed: %v", err)
	}
	defer hostConn.Close()

	playerConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?playerId=p1", nil)
	if err != nil {
		t.Fatalf("player dial failed: %v", err)
	}
	defer playerConn.Close()

	// --- Round 1: open with a 1-second countdown ---

	if err := hostConn.WriteJSON(map[string]string{"type": "ready"}); err != nil {
		t.Fatal(err)
	}
	waitForState(t, hostConn, "ready")

	before := time.Now()

	if err := hostConn.WriteJSON(map[string]any{"type": "open", "seconds": 1}); err != nil {
		t.Fatal(err)
	}

	countdownUpdate := waitForState(t, hostConn, "countdown")

	if countdownUpdate.Lobby.CountdownEndsAt == nil {
		t.Fatal("expected CountdownEndsAt to be set during countdown")
	}

	if countdownUpdate.Lobby.CountdownEndsAt.Before(before.Add(500 * time.Millisecond)) {
		t.Fatalf("expected countdown to end roughly 1s out, got %v (started %v)", countdownUpdate.Lobby.CountdownEndsAt, before)
	}

	// The countdown should resolve to "open" on its own, without any
	// further client message.
	waitForState(t, hostConn, "open")

	if err := playerConn.WriteJSON(map[string]string{"type": "buzz", "playerId": "p1"}); err != nil {
		t.Fatal(err)
	}

	waitForState(t, playerConn, "locked")

	// --- Close round 1, check scoring, start round 2 ---

	if err := hostConn.WriteJSON(map[string]string{"type": "next_round"}); err != nil {
		t.Fatal(err)
	}

	afterRound1 := waitForState(t, hostConn, "ready")

	if afterRound1.Lobby.RoundNumber != 2 {
		t.Fatalf("expected round number 2, got %d", afterRound1.Lobby.RoundNumber)
	}

	if len(afterRound1.Lobby.Scores) != 1 || afterRound1.Lobby.Scores[0].PlayerID != "p1" || afterRound1.Lobby.Scores[0].Points != 1 {
		t.Fatalf("expected p1 to have 1 point after round 1, got %+v", afterRound1.Lobby.Scores)
	}

	if len(afterRound1.Lobby.History) != 1 || afterRound1.Lobby.History[0].WinnerID != "p1" {
		t.Fatalf("expected round 1 in history with p1 as winner, got %+v", afterRound1.Lobby.History)
	}

	// --- Round 2: instant open (no countdown), p2 wins this time ---

	if err := hostConn.WriteJSON(map[string]any{"type": "open", "seconds": 0}); err != nil {
		t.Fatal(err)
	}
	waitForState(t, hostConn, "open")

	if err := playerConn.WriteJSON(map[string]string{"type": "buzz", "playerId": "p2"}); err != nil {
		t.Fatal(err)
	}

	waitForState(t, hostConn, "locked")

	if err := hostConn.WriteJSON(map[string]string{"type": "next_round"}); err != nil {
		t.Fatal(err)
	}

	afterRound2 := waitForState(t, hostConn, "ready")

	if afterRound2.Lobby.RoundNumber != 3 {
		t.Fatalf("expected round number 3, got %d", afterRound2.Lobby.RoundNumber)
	}

	if len(afterRound2.Lobby.Scores) != 2 {
		t.Fatalf("expected both players to have a score, got %+v", afterRound2.Lobby.Scores)
	}

	for _, s := range afterRound2.Lobby.Scores {
		if s.Points != 1 {
			t.Fatalf("expected each player to have exactly 1 point, got %+v", s)
		}
	}

	if len(afterRound2.Lobby.History) != 2 {
		t.Fatalf("expected 2 rounds in history, got %d", len(afterRound2.Lobby.History))
	}
}

// TestWebsocketNonHostCannotStartNextRound confirms next_round is
// host-gated just like ready/open.
func TestWebsocketNonHostCannotStartNextRound(t *testing.T) {
	router := newTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	rec := doJSON(t, router, http.MethodPost, "/api/lobbies", map[string]any{
		"name": "Quiz night", "hostId": "host1", "public": true,
	})

	var created dto.LobbyResponse
	json.Unmarshal(rec.Body.Bytes(), &created)

	doJSON(t, router, http.MethodPost, "/api/lobbies/"+created.ID+"/join", map[string]any{
		"id": "p1", "name": "Alice",
	})

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/lobbies/" + created.ID + "/ws"

	hostConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?playerId=host1", nil)
	if err != nil {
		t.Fatalf("host dial failed: %v", err)
	}
	defer hostConn.Close()

	playerConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?playerId=p1", nil)
	if err != nil {
		t.Fatalf("player dial failed: %v", err)
	}
	defer playerConn.Close()

	if err := hostConn.WriteJSON(map[string]string{"type": "ready"}); err != nil {
		t.Fatal(err)
	}
	waitForState(t, hostConn, "ready")

	if err := hostConn.WriteJSON(map[string]any{"type": "open", "seconds": 0}); err != nil {
		t.Fatal(err)
	}
	waitForState(t, hostConn, "open")

	if err := playerConn.WriteJSON(map[string]string{"type": "buzz", "playerId": "p1"}); err != nil {
		t.Fatal(err)
	}
	waitForState(t, playerConn, "locked")

	// p1 (not the host) tries to start the next round.
	if err := playerConn.WriteJSON(map[string]string{"type": "next_round", "playerId": "p1"}); err != nil {
		t.Fatal(err)
	}

	playerConn.SetReadDeadline(time.Now().Add(2 * time.Second))

	var msg testOutbound
	if err := playerConn.ReadJSON(&msg); err != nil {
		t.Fatalf("expected an error message, got read error: %v", err)
	}

	if msg.Type != "error" {
		t.Fatalf("expected an error message for a non-host next_round, got %+v", msg)
	}
}

type testOutbound struct {
	Type  string             `json:"type"`
	Lobby *dto.LobbyResponse `json:"lobby,omitempty"`
	Error string             `json:"error,omitempty"`
}

func readUpdate(t *testing.T, conn *websocket.Conn) testOutbound {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	var msg testOutbound

	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	return msg
}

// waitForState reads messages off conn until it sees a lobby_update whose
// state matches want, skipping unrelated broadcasts triggered by the
// other connected client.
func waitForState(t *testing.T, conn *websocket.Conn, want string) testOutbound {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		msg := readUpdate(t, conn)

		if msg.Type == "lobby_update" && msg.Lobby != nil && msg.Lobby.State == want {
			return msg
		}
	}

	t.Fatalf("never observed state %q", want)

	return testOutbound{}
}
