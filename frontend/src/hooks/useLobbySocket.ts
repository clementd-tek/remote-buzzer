import { useCallback, useEffect, useRef, useState } from "react";
import type { ClientMessage, Lobby, ServerMessage } from "../types/lobby";

function wsUrl(lobbyId: string, playerId: string): string {
  const base = import.meta.env.VITE_API_BASE_URL as string | undefined;
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const origin = base ? base.replace(/^https?/, protocol) : `${protocol}://${window.location.host}`;

  return `${origin}/api/lobbies/${lobbyId}/ws?playerId=${encodeURIComponent(playerId)}`;
}

export type ConnectionStatus = "connecting" | "open" | "reconnecting" | "closed";

interface UseLobbySocketResult {
  lobby: Lobby | null;
  status: ConnectionStatus;
  lastError: string | null;
  send: (message: ClientMessage) => void;
}

const RECONNECT_DELAY_MS = 1500;

/**
 * Owns the websocket connection for a single lobby. Keeps the latest
 * lobby_update in state, surfaces the most recent server-side error (e.g.
 * "only the host can open the buzzer"), and reconnects automatically if
 * the connection drops unexpectedly.
 */
export function useLobbySocket(lobbyId: string, playerId: string | null): UseLobbySocketResult {
  const [lobby, setLobby] = useState<Lobby | null>(null);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const [lastError, setLastError] = useState<string | null>(null);

  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const closedByClient = useRef(false);

  useEffect(() => {
    if (!playerId) return;

    closedByClient.current = false;

    function connect() {
      setStatus((prev) => (prev === "open" ? prev : "connecting"));

      const socket = new WebSocket(wsUrl(lobbyId, playerId!));
      socketRef.current = socket;

      socket.onopen = () => setStatus("open");

      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as ServerMessage;

          if (message.type === "lobby_update") {
            setLobby(message.lobby);
            setLastError(null);
          } else if (message.type === "error") {
            setLastError(message.error);
          }
        } catch {
          // Ignore malformed frames rather than crashing the UI.
        }
      };

      socket.onclose = () => {
        if (closedByClient.current) {
          setStatus("closed");
          return;
        }

        setStatus("reconnecting");
        reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY_MS);
      };

      socket.onerror = () => {
        socket.close();
      };
    }

    connect();

    return () => {
      closedByClient.current = true;

      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
      }

      socketRef.current?.close();
    };
  }, [lobbyId, playerId]);

  const send = useCallback((message: ClientMessage) => {
    const socket = socketRef.current;

    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(message));
    }
  }, []);

  return { lobby, status, lastError, send };
}
