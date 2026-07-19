# Le Buzzer — frontend

Vite + React + TypeScript client for the real-time buzzer app. Pairs with
the Go backend in `../backend`.

## Develop locally

```bash
npm install
npm run dev
```

Runs on `http://localhost:5173`. The dev server proxies `/api/*`
(REST and websocket) to a backend running on `http://localhost:8080`
(see `vite.config.ts`), so just start the backend separately:

```bash
cd ../backend
go run ./cmd/server/main.go
```

No `.env` needed for this setup — REST calls and the websocket both use
relative `/api/...` paths by default (see `src/api/client.ts`). Set
`VITE_API_BASE_URL` only if you want to point at a backend running
somewhere other than what the dev proxy / production nginx already
routes to.

## Build

```bash
npm run build   # type-checks with tsc, then builds to dist/
npm run preview # serve the production build locally
```

## Project layout

```
src/
  api/        REST client + localStorage identity helpers
  components/ Buzzer, forms, roster, status indicators
  hooks/      useLobbySocket — the websocket connection
  pages/      HomePage, LobbyPage
  types/      Shared types mirroring the backend DTOs
```

## How identity works

There's no login. Creating a lobby generates a random host id, stored in
`localStorage` under the lobby's id. Joining a lobby does the same for a
player id. Opening the lobby's link in a different browser (or after
clearing storage) starts a fresh join.

## Docker

`Dockerfile` builds the app and serves the static output via nginx. See
the root `docker-compose.yml` for how it fits with the backend and the
load-balancing nginx in front of everything.

## Tests

```bash
npm run test
```

Runs against a real backend on `http://localhost:8080` (see
`.env.test`) rather than mocking fetch/websocket — start the backend
first (`cd ../backend && go run ./cmd/server/main.go`).

## A gotcha this app used to hit

If the browser's page origin doesn't match what the backend's websocket
upgrade handler allows, REST calls keep working fine (they're proxied
server-side, invisible to browser CORS) while the websocket silently
fails to connect — so the UI looks like it's frozen on stale data with
only a small "reconnecting" indicator as a clue. The backend now always
allows any `localhost`/`127.0.0.1` origin regardless of port (see
`internal/originpolicy` on the backend) specifically because Vite picks
a different port than 5173 whenever 5173 is already taken. `strictPort:
true` in `vite.config.ts` also makes that port collision fail loudly
instead of silently drifting. The frontend additionally polls over REST
as a fallback whenever the websocket isn't connected, so even a genuine
network issue won't leave the UI permanently stale.
