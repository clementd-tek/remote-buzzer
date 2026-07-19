import type { Lobby, Player } from "../types/lobby";

// Relative by default: the Vite dev server proxies /api to the backend
// (see vite.config.ts), and in production nginx does the same in front
// of the built static files. Set VITE_API_BASE_URL to override (e.g. to
// point at a different backend while developing).
const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";

class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (!response.ok) {
    const text = await response.text();
    throw new ApiError(response.status, text || response.statusText);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export interface CreateLobbyInput {
  name: string;
  hostId: string;
  public: boolean;
}

export function createLobby(input: CreateLobbyInput): Promise<Lobby> {
  return request<Lobby>("/api/lobbies", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function listPublicLobbies(): Promise<Lobby[]> {
  return request<Lobby[]>("/api/lobbies");
}

export function getLobby(id: string): Promise<Lobby> {
  return request<Lobby>(`/api/lobbies/${id}`);
}

export interface JoinLobbyInput {
  id: string;
  name: string;
}

export function joinLobby(lobbyId: string, input: JoinLobbyInput): Promise<Player> {
  return request<Player>(`/api/lobbies/${lobbyId}/join`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export { ApiError };
