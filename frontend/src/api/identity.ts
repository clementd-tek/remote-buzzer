// There's no auth in this app: a browser's role in a lobby (host, or a
// joined player) is remembered in localStorage, scoped per lobby id. This
// is deliberately lightweight — good enough for "share a link with
// friends", not meant to survive clearing browser data or switching
// devices.

function hostKey(lobbyId: string) {
  return `buzzer:host:${lobbyId}`;
}

function playerKey(lobbyId: string) {
  return `buzzer:player:${lobbyId}`;
}

// crypto.randomUUID() is only available in secure contexts (HTTPS or
// localhost). When the app is served over plain HTTP on a LAN IP we fall
// back to a Math.random-based v4 UUID. It is not cryptographically strong
// but is more than sufficient for ephemeral lobby/player IDs.
export function newId(): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return crypto.randomUUID();
  }

  // RFC-4122 v4 UUID fallback
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

export function rememberHost(lobbyId: string, hostId: string) {
  localStorage.setItem(hostKey(lobbyId), hostId);
}

export function getRememberedHostId(lobbyId: string): string | null {
  return localStorage.getItem(hostKey(lobbyId));
}

export interface RememberedPlayer {
  id: string;
  name: string;
}

export function rememberPlayer(lobbyId: string, player: RememberedPlayer) {
  localStorage.setItem(playerKey(lobbyId), JSON.stringify(player));
}

export function getRememberedPlayer(lobbyId: string): RememberedPlayer | null {
  const raw = localStorage.getItem(playerKey(lobbyId));

  if (!raw) return null;

  try {
    return JSON.parse(raw) as RememberedPlayer;
  } catch {
    return null;
  }
}
