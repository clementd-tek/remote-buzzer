# Contributing to Remote Buzzer

Thank you for taking the time to contribute! This document covers everything you need to go from zero to an open pull request.

---

## Table of contents

- [Code of conduct](#code-of-conduct)
- [How to report a bug](#how-to-report-a-bug)
- [How to suggest a feature](#how-to-suggest-a-feature)
- [Development setup](#development-setup)
- [Making a change](#making-a-change)
- [Pull request checklist](#pull-request-checklist)
- [Code style](#code-style)
- [Commit messages](#commit-messages)

---

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). By participating you agree to abide by its terms.

---

## How to report a bug

1. Search [existing issues](../../issues) first — someone may have already reported it.
2. If not, open a **Bug report** issue using the template. Include:
   - What you did
   - What you expected to happen
   - What actually happened
   - Your environment (OS, browser, Docker version)
   - Relevant logs (`docker compose logs backend`)

---

## How to suggest a feature

Open a **Feature request** issue. Describe the problem you're trying to solve, not just the solution you have in mind — this helps discussion stay open.

---

## Development setup

### Requirements

| Tool | Minimum version |
|------|----------------|
| Go | 1.22 |
| Node.js | 20 |
| Docker | 24 |
| Docker Compose | v2 |

### First-time setup

```bash
git clone https://github.com/clementd-tek/remote-buzzer.git
cd remote-buzzer
cp .env.example .env
```

### Running the full stack

```bash
docker compose up --build
# → http://localhost:8080
```

### Running services individually (faster iteration)

**Backend only:**
```bash
cd backend
go run ./cmd/server
# → http://localhost:8080/api
```

**Frontend only** (proxies `/api` to the backend above):
```bash
cd frontend
npm install
npm run dev
# → http://localhost:5173
```

---

## Making a change

1. **Fork** the repository and clone your fork.
2. Create a **feature branch** from `main`:
   ```bash
   git checkout -b feat/short-description
   # or: fix/short-description
   ```
3. Make your changes (see [Code style](#code-style) below).
4. **Write or update tests** for your change.
5. Run the test suites:
   ```bash
   # Backend
   cd backend && go test ./...

   # Frontend
   cd frontend && npm test
   ```
6. Commit following the [commit message convention](#commit-messages).
7. Push and open a **pull request** against `main`.

---

## Pull request checklist

Before marking your PR as ready for review:

- [ ] Tests pass locally (`go test ./...` and `npm test`)
- [ ] New behaviour is covered by tests
- [ ] The PR description explains *what* changed and *why*
- [ ] `FRONTEND_ORIGINS` / env changes are reflected in `.env.example` and `README.md`
- [ ] No secrets, credentials, or personal data are committed

---

## Code style

### Go (backend)

- Standard `gofmt` formatting — run `gofmt -w .` before committing, or configure your editor to do it on save.
- Exported symbols have doc comments.
- Errors are returned, not logged-and-swallowed at the call site.
- New packages go under `internal/` — nothing in `internal/` is considered a public API.

### TypeScript (frontend)

- The project uses `oxlint` for linting: `npm run lint`
- Prefer named exports over default exports for components.
- Keep components focused — if a file is getting long, split it.
- Hooks live in `src/hooks/`, pure API calls in `src/api/`.

### General

- Prefer clarity over cleverness.
- Delete commented-out code — use git history instead.
- Keep PRs focused. One logical change per PR is easier to review and revert.

---

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>

[optional body]

[optional footer]
```

Common types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`.

Examples:
```
feat(lobby): add private lobby support
fix(originpolicy): allow RFC-1918 addresses for LAN access
docs: add HTTPS deployment guide to README
test(ws): cover CheckOrigin with LAN IP cases
```

The summary line should complete the sentence *"If applied, this commit will …"* — so `add private lobby support`, not `added` or `adding`.
