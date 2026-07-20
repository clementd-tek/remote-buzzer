import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { test, expect, vi, beforeEach } from "vitest";
import App from "./App";
import { createLobby, joinLobby } from "./api/client";
import { rememberHost, rememberPlayer } from "./api/identity";
import type { Lobby } from "./types/lobby";
import * as socketHook from "./hooks/useLobbySocket";

// The real websocket flow is already covered thoroughly by the backend's
// own integration tests (internal/api/router_test.go), which exercise a
// genuine websocket connection end to end. Here we mock the hook instead
// of using a real socket: this environment's WebSocket implementation
// (Node's undici running under jsdom) never dispatches open/close/error
// events, only message events — a documented tooling limitation, not
// something a real browser hits — so any test gating on "connected"
// would hang regardless of how it's written. Mocking keeps these tests
// focused on what they should verify: does the UI render the right
// thing for a given lobby state.
vi.mock("./hooks/useLobbySocket", () => ({
  useLobbySocket: vi.fn(),
}));

const send = vi.fn();

function mockSocket(lobby: Lobby) {
  vi.mocked(socketHook.useLobbySocket).mockReturnValue({
    lobby,
    status: "open",
    lastError: null,
    send,
  });
}

function baseLobby(overrides: Partial<Lobby>): Lobby {
  return {
    id: "lobby-1",
    name: "Test Lobby",
    public: true,
    state: "waiting",
    hostId: "host1",
    playerCount: 1,
    players: [{ id: "p1", name: "Alice" }],
    roundNumber: 1,
    scores: [],
    history: [],
    ...overrides,
  };
}

beforeEach(() => {
  send.mockClear();
});

test("host sees the countdown number and the round number", async () => {
  const created = await createLobby({ name: "Countdown Test", hostId: "host1", public: true });
  await joinLobby(created.id, { id: "p1", name: "Alice" });
  rememberHost(created.id, "host1");

  mockSocket(
    baseLobby({
      id: created.id,
      state: "countdown",
      countdownEndsAt: new Date(Date.now() + 2000).toISOString(),
      roundNumber: 1,
    }),
  );

  render(
    <MemoryRouter initialEntries={[`/lobby/${created.id}`]}>
      <App />
    </MemoryRouter>,
  );

  await screen.findAllByText(/compte à rebours/i);
  expect(screen.getAllByText("Manche 1").length).toBeGreaterThan(0);
});

test("host sees a next-round button and the scoreboard once a round is locked", async () => {
  const created = await createLobby({ name: "Locked Test", hostId: "host1", public: true });
  await joinLobby(created.id, { id: "p1", name: "Alice" });
  rememberHost(created.id, "host1");

  mockSocket(
    baseLobby({
      id: created.id,
      state: "locked",
      winner: { playerId: "p1", time: new Date().toISOString() },
      roundNumber: 1,
      scores: [{ playerId: "p1", name: "Alice", points: 1 }],
      history: [
        { round: 1, winnerId: "p1", winnerName: "Alice", points: 1, closedAt: new Date().toISOString() },
      ],
    }),
  );

  render(
    <MemoryRouter initialEntries={[`/lobby/${created.id}`]}>
      <App />
    </MemoryRouter>,
  );

  await screen.findByText(/a buzzé en premier/i);
  expect(screen.getAllByText("Alice", { exact: false }).length).toBeGreaterThan(0);
  expect(screen.getByText("1 pt")).toBeInTheDocument();

  const nextRoundButton = screen.getByRole("button", { name: /manche suivante/i });
  nextRoundButton.click();

  expect(send).toHaveBeenCalledWith({ type: "next_round" });
});

test("host open button sends the shared countdown duration", async () => {
  const created = await createLobby({ name: "Open Test", hostId: "host1", public: true });
  await joinLobby(created.id, { id: "p1", name: "Alice" });
  rememberHost(created.id, "host1");

  mockSocket(baseLobby({ id: created.id, state: "ready" }));

  render(
    <MemoryRouter initialEntries={[`/lobby/${created.id}`]}>
      <App />
    </MemoryRouter>,
  );

  const openButton = await screen.findByRole("button", { name: /lancer le buzzer/i });
  openButton.click();

  expect(send).toHaveBeenCalledWith({ type: "open", seconds: 3 });
});

test("player sees the countdown number on the buzzer, then win framing once locked in their favor", async () => {
  const created = await createLobby({ name: "Player View Test", hostId: "host1", public: true });
  await joinLobby(created.id, { id: "p1", name: "Alice" });
  rememberPlayer(created.id, { id: "p1", name: "Alice" });

  mockSocket(
    baseLobby({
      id: created.id,
      state: "locked",
      winner: { playerId: "p1", time: new Date().toISOString() },
      scores: [{ playerId: "p1", name: "Alice", points: 1 }],
    }),
  );

  render(
    <MemoryRouter initialEntries={[`/lobby/${created.id}`]}>
      <App />
    </MemoryRouter>,
  );

  await screen.findByText(/gagné/i);
  // The player's own scoreboard should also reflect the point.
  expect(screen.getByText("1 pt")).toBeInTheDocument();
});
