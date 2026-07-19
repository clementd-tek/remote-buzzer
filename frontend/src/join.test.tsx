import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { test, expect } from "vitest";
import App from "./App";
import { createLobby } from "./api/client";

// Regression test for a real bug: a guest could join a lobby
// successfully over REST, yet the roster kept showing "nobody has
// joined" because the websocket's Origin header didn't match the
// backend's allowed-origins list, so the live update never arrived and
// the UI was stuck on the stale pre-join snapshot. See
// internal/originpolicy on the backend for the fix.
test("a fresh guest joining an existing lobby sees themselves in the roster", async () => {
  const lobby = await createLobby({ name: "Join Test", hostId: "host-abc", public: true });
  const user = userEvent.setup();

  const { unmount } = render(
    <MemoryRouter initialEntries={[`/lobby/${lobby.id}`]}>
      <App />
    </MemoryRouter>,
  );

  const nameInput = await screen.findByPlaceholderText("Ton prénom");
  await user.type(nameInput, "Bob");
  await user.click(screen.getByRole("button", { name: /rejoindre/i }));

  await waitFor(
    () => {
      expect(screen.getByText("Bob")).toBeInTheDocument();
    },
    { timeout: 8000 },
  );

  // Explicitly close the websocket (and its reconnect timer) before the
  // test ends, rather than leaving it running in the background.
  unmount();
}, 10000);
