export type LobbyState = "waiting" | "ready" | "open" | "locked";

export interface Player {
  id: string;
  name: string;
}

export interface Winner {
  playerId: string;
  time: string;
}

export interface Lobby {
  id: string;
  name: string;
  public: boolean;
  state: LobbyState;
  hostId: string;
  playerCount: number;
  players: Player[];
  winner?: Winner;
}

/** Messages the client can send over the lobby websocket. */
export type ClientMessage =
  | { type: "ready" }
  | { type: "open" }
  | { type: "buzz"; playerId: string };

/** Messages the server pushes down the lobby websocket. */
export type ServerMessage =
  | { type: "lobby_update"; lobby: Lobby }
  | { type: "error"; error: string };
