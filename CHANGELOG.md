# Changelog

All notable changes to this project will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). This project uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- `FRONTEND_ORIGINS` env var (replaces `FRONTEND_ORIGIN`) — supports a comma-separated list of allowed CORS / WebSocket origins
- `IsPrivateNetwork` origin check — RFC-1918 LAN addresses (`192.168.x.x`, `10.x.x.x`, `172.16–31.x.x`) are now always allowed without explicit configuration, enabling LAN access out of the box
- `IsAllowed` unified function in `originpolicy` — used by both the CORS middleware and the WebSocket upgrader to prevent drift between the two checks
- `crypto.randomUUID()` fallback in `identity.ts` — the app now works over plain HTTP on a LAN (insecure context) using a `Math.random`-based UUID v4 when the native API is unavailable

### Changed
- `docker-compose.yml` now reads configuration from `.env` via `env_file`
- `config.go` parses origins at load time into `[]string`, removing per-callsite splitting logic

### Fixed
- WebSocket connections from LAN IPs were silently rejected because `hub.go` only called `IsLocal`, not the updated `IsAllowed`
- Lobby creation failed silently on LAN devices because `crypto.randomUUID()` threw in an insecure context before the POST request was ever made

---

## [0.1.0] — 2026-07-22

### Added
- Initial release
- Lobby creation, listing, and joining
- Real-time WebSocket buzzer with server-side arbitration
- Countdown mode (configurable seconds before buzzer opens)
- Scoreboard with configurable points per round
- Public / private lobby toggle
- Valkey-backed lobby persistence with TTL eviction
- Horizontal backend scaling via `docker compose deploy.replicas`
- nginx reverse proxy with WebSocket support and `ip_hash` sticky sessions
- GitHub Actions CI: Go tests + frontend tests on every push/PR
- GitHub Actions CD: Docker images pushed to GHCR on merge to `main`
