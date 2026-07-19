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

export function newId(): string {
  return crypto.randomUUID();
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
