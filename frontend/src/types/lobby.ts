export type LobbyState = "waiting" | "ready" | "countdown" | "open" | "locked";

export interface Player {
  id: string;
  name: string;
}

export interface Winner {
  playerId: string;
  time: string;
}

export interface Score {
  playerId: string;
  name: string;
  points: number;
}

export interface RoundResult {
  round: number;
  winnerId: string;
  winnerName: string;
  points: number;
  closedAt: string;
}

export interface LobbySettings {
  pointsPerRound: number;
  countdownSeconds: number;
}

export const DEFAULT_SETTINGS: LobbySettings = {
  pointsPerRound: 1,
  countdownSeconds: 3,
};

export const MAX_POINTS_PER_ROUND = 20;
export const MAX_COUNTDOWN_SECONDS = 30;

export interface Lobby {
  id: string;
  name: string;
  public: boolean;
  state: LobbyState;
  hostId: string;
  playerCount: number;
  players: Player[];
  winner?: Winner;
  roundNumber: number;
  countdownEndsAt?: string;
  scores: Score[];
  history: RoundResult[];
  settings: LobbySettings;
}

/** Messages the client can send over the lobby websocket. */
export type ClientMessage =
  | { type: "ready" }
  | { type: "open" }
  | { type: "buzz"; playerId: string }
  | { type: "next_round" }
  | { type: "settings"; pointsPerRound?: number; countdownSeconds?: number };

/** Messages the server pushes down the lobby websocket. */
export type ServerMessage =
  | { type: "lobby_update"; lobby: Lobby }
  | { type: "error"; error: string };
