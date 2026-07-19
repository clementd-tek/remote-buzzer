#!/bin/bash

# Quick curl test suite for the remote-buzzer backend.
# Run the backend locally first: go run ./cmd/server/main.go
# Then: bash test.sh

set -e

BASE_URL="http://localhost:8080/api"
HOST_ID="host-$(date +%s)"

echo "=== Creating a public lobby ==="
LOBBY=$(curl -s -X POST "$BASE_URL/lobbies" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"Test Quiz\", \"hostId\": \"$HOST_ID\", \"public\": true}")

LOBBY_ID=$(echo "$LOBBY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo "Created lobby: $LOBBY_ID"
echo "$LOBBY" | grep -o '"state":"[^"]*'

echo ""
echo "=== Listing public lobbies (should see ours) ==="
curl -s -X GET "$BASE_URL/lobbies" | grep -o '"name":"[^"]*'

echo ""
echo "=== Getting the lobby details ==="
curl -s -X GET "$BASE_URL/lobbies/$LOBBY_ID" | grep -o '"state":"[^"]*'

echo ""
echo "=== Joining as player 1 ==="
curl -s -X POST "$BASE_URL/lobbies/$LOBBY_ID/join" \
  -H "Content-Type: application/json" \
  -d '{"id": "p1", "name": "Alice"}' | grep -o '"id":"[^"]*'

echo ""
echo "=== Joining as player 2 ==="
curl -s -X POST "$BASE_URL/lobbies/$LOBBY_ID/join" \
  -H "Content-Type: application/json" \
  -d '{"id": "p2", "name": "Bob"}' | grep -o '"id":"[^"]*'

echo ""
echo "=== Host readies the lobby (REST call) ==="
curl -s -X POST "$BASE_URL/lobbies/$LOBBY_ID/ready" \
  -H "Content-Type: application/json" \
  -d "{\"hostId\": \"$HOST_ID\"}" 2>&1 | head -1 || echo "(Not yet wired to REST; use websocket instead)"

echo ""
echo "=== Health check ==="
curl -s -X GET "http://localhost:8080/healthz"

echo ""
echo ""
echo "For real-time testing (ready, open, buzz), use websocat or a browser websocket client:"
echo "  websocat 'ws://localhost:8080/api/lobbies/$LOBBY_ID/ws?playerId=host-123'"
echo ""
echo "Then send JSON messages like:"
echo "  {\"type\": \"ready\"}"
echo "  {\"type\": \"open\"}"
echo "  {\"type\": \"buzz\", \"playerId\": \"p1\"}"
