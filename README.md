# 🔴 Remote Buzzer

A real-time, network buzzer app inspired by TV game shows like *Questions pour un champion*. Create a lobby, share the link — the server decides who buzzed first.

```
Host creates lobby → shares link → players join → host opens round → fastest finger wins
```

**Stack:** Go · React · WebSockets · Valkey · nginx · Docker

---

## Features

- **Instant buzz arbitration** — the server timestamps every buzz server-side; no client clock drift can cheat the result
- **Real-time sync** — all players see the winner the moment it's decided, via WebSocket push
- **Lobby browser** — public lobbies are listed on the home page; private lobbies are link-only
- **Countdown mode** — optional server-driven countdown before the buzzer opens (3, 2, 1 …)
- **Scoreboard** — configurable points per round, persistent across rounds within a session
- **Scales horizontally** — multiple backend replicas share state through Valkey; nginx uses `ip_hash` to pin WebSocket connections

---

## Architecture

```
Browser
  │  HTTP + WebSocket
  ▼
nginx  ──────────────────────────────┐
  │  /api/*  →  backend (Go/Chi)    │  /  →  frontend (React, static)
  ▼                                  ▼
backend:8080              frontend nginx :80
  │
  └── Valkey (lobby directory + TTL eviction)
```

| Service    | Role |
|------------|------|
| `backend`  | REST API + WebSocket hub, written in Go with Chi |
| `frontend` | React SPA built with Vite, served as static files by nginx |
| `nginx`    | Reverse proxy — routes `/api/` to backend, everything else to frontend |
| `valkey`   | In-memory cache for lobby state; backend falls back to in-process memory if unreachable |

---

## Quick start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) ≥ 24
- [Docker Compose](https://docs.docker.com/compose/install/) v2

### Run locally

```bash
git clone https://github.com/clementd-tek/remote-buzzer.git
cd remote-buzzer
cp .env.example .env
docker compose up --build
```

Open [http://localhost:8080](http://localhost:8080).

### Access from other devices on your LAN

The app works over plain HTTP on a local network. Just open `http://<your-machine-ip>:8080` from any device on the same network — no extra configuration needed.

> **Note:** `crypto.randomUUID()` requires HTTPS in most browsers. The app includes an automatic fallback for HTTP/LAN use, but for production deployments HTTPS is strongly recommended (see [Deployment](#deployment)).

---

## Configuration

All configuration lives in `.env`. Copy `.env.example` to get started:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `FRONTEND_ORIGINS` | `http://localhost:8080` | Comma-separated list of allowed CORS / WebSocket origins. Private LAN IPs (`192.168.x.x`, `10.x.x.x`, `172.16–31.x.x`) and `localhost` are **always** allowed regardless of this value. |
| `VALKEY_ADDR` | `valkey:6379` | Valkey/Redis address. Leave empty to run purely in-memory. |
| `VALKEY_PASSWORD` | _(empty)_ | Valkey password, if any. |
| `LOBBY_TTL` | `6h` | How long an idle lobby is kept before cleanup. |
| `LOBBY_CLEANUP_INTERVAL` | `10m` | How often the cleanup loop runs. |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown drain window. |

**Multiple origins example:**
```env
FRONTEND_ORIGINS="https://buzzer.example.com,https://admin.example.com"
```

---

## Development

### Backend (Go)

```bash
cd backend
go run ./cmd/server          # start with in-memory store (no Valkey needed)
go test ./...                # run all tests
```

Requires Go ≥ 1.22.

### Frontend (React + Vite)

```bash
cd frontend
npm install
npm run dev                  # dev server on :5173, proxies /api → localhost:8080
npm test                     # vitest unit tests
```

Requires Node ≥ 20.

The Vite dev server proxies `/api/*` to `http://localhost:8080`, so you can run `go run ./cmd/server` alongside `npm run dev` without Docker.

### Full stack with Docker Compose (watch mode)

```bash
docker compose up --build --watch
```

---

## Project structure

```
remote-buzzer/
├── backend/
│   ├── cmd/server/          # main entrypoint
│   └── internal/
│       ├── api/             # Chi router, handlers, DTOs
│       ├── app/             # wires everything together
│       ├── cache/           # Valkey client
│       ├── config/          # env-based config
│       ├── lobby/           # lobby domain model, manager, service
│       ├── originpolicy/    # unified CORS + WS origin check
│       └── ws/              # WebSocket hub and client pumps
├── frontend/
│   └── src/
│       ├── api/             # fetch client + identity helpers
│       ├── components/      # reusable UI components
│       ├── hooks/           # useLobbySocket (WS + REST fallback)
│       ├── pages/           # HomePage, LobbyPage
│       └── types/           # shared TypeScript types
├── nginx/
│   └── nginx.conf           # top-level reverse proxy
├── docker-compose.yml
└── .env.example
```

---

## API reference

All endpoints are under `/api/`.

### Lobbies

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/lobbies` | Create a lobby |
| `GET` | `/api/lobbies` | List public lobbies |
| `GET` | `/api/lobbies/:id` | Get a single lobby |
| `POST` | `/api/lobbies/:id/join` | Join a lobby as a player |
| `GET` | `/api/lobbies/:id/ws` | Upgrade to WebSocket |

### WebSocket messages

**Server → client:**

```jsonc
{ "type": "lobby_update", "lobby": { ... } }
{ "type": "error", "error": "only the host can open the buzzer" }
```

**Client → server:**

```jsonc
{ "type": "ready" }                         // host: mark lobby as ready
{ "type": "open" }                          // host: open the buzzer
{ "type": "buzz", "playerId": "..." }       // player: buzz
{ "type": "next_round" }                    // host: advance to next round
{ "type": "settings", "pointsPerRound": 3, "countdownSeconds": 5 }
```

---

## Deployment

### Docker Compose (self-hosted)

```bash
cp .env.example .env
# edit .env — set FRONTEND_ORIGINS to your domain
docker compose pull   # if using pre-built images from GHCR
docker compose up -d
```

### HTTPS (recommended for production)

Running behind HTTPS restores full browser security context:
- `crypto.randomUUID()` works natively
- WebSocket uses `wss://` (encrypted)

The simplest approach is to put a TLS-terminating reverse proxy (Caddy, Traefik, or nginx with Certbot) in front of the `nginx` service. Point it at port 8080 and set `FRONTEND_ORIGINS` to your `https://` domain.

### Scaling backends

Edit `docker-compose.yml`:

```yaml
backend:
  deploy:
    replicas: 3
```

nginx uses `ip_hash` to pin each client to one backend instance (needed because WebSocket connections are stateful). Lobby metadata is shared across replicas via Valkey.

---

## Contributing

Contributions are welcome — bug reports, feature ideas, and PRs alike.

1. **Fork** the repo and create a branch: `git checkout -b feat/my-feature`
2. **Make your changes.** Run tests before pushing:
   ```bash
   cd backend && go test ./...
   cd frontend && npm test
   ```
3. **Open a pull request** against `main`. Fill in the PR template — describe what changed and why.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for the full guidelines, and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.

### Good first issues

Look for issues tagged [`good first issue`](../../issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) — these are well-scoped and include enough context to get started without deep knowledge of the codebase.

---

## License

[BSD 3-Clause](LICENSE) © 2026 Clément DEVAUX
