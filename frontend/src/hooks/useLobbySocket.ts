import { useCallback, useEffect, useRef, useState } from "react";
import { getLobby } from "../api/client";
import { DEFAULT_SETTINGS, type ClientMessage, type Lobby, type ServerMessage } from "../types/lobby";

function wsUrl(lobbyId: string, playerId: string): string {
  const base = import.meta.env.VITE_API_BASE_URL as string | undefined;
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const origin = base ? base.replace(/^https?/, protocol) : `${protocol}://${window.location.host}`;

  return `${origin}/api/lobbies/${lobbyId}/ws?playerId=${encodeURIComponent(playerId)}`;
}

/** Ensure fields added in recent versions are always present, even if a
 * browser session was opened against an older server. */
function normaliseLobby(raw: Lobby): Lobby {
  return {
    ...raw,
    settings: raw.settings ?? DEFAULT_SETTINGS,
    scores: raw.scores ?? [],
    history: raw.history ?? [],
    players: raw.players ?? [],
  };
}

export type ConnectionStatus = "connecting" | "open" | "reconnecting" | "closed";

interface UseLobbySocketResult {
  lobby: Lobby | null;
  status: ConnectionStatus;
  lastError: string | null;
  send: (message: ClientMessage) => void;
}

const RECONNECT_DELAY_MS = 1500;
const FALLBACK_POLL_MS = 3000;

/**
 * Owns the websocket connection for a single lobby. Keeps the latest
 * lobby_update in state, surfaces the most recent server-side error (e.g.
 * "only the host can open the buzzer"), and reconnects automatically if
 * the connection drops unexpectedly.
 *
 * While the socket isn't open (initial connect, or a drop that hasn't
 * recovered yet), this also polls the REST endpoint every few seconds as
 * a fallback. Without that, a lobby that never manages to open a
 * websocket (misconfigured origin, restrictive network, ...) would be
 * stuck showing whatever snapshot happened to be loaded before the
 * connection attempt — including a roster missing whoever just joined —
 * with only a small status indicator hinting that anything's wrong.
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
            setLobby(normaliseLobby(message.lobby));
            setLastError(null);
          } else if (message.type === "error") {
            setLastError(message.error);
          }
        } catch {
          // Ignore malformed frames rather than crashing the UI.
        }
      };

      socket.onclose = (event) => {
        if (closedByClient.current) {
          setStatus("closed");
          return;
        }

        if (!event.wasClean) {
          // Helps diagnose the real cause (e.g. a rejected origin) from
          // the browser console instead of just "it's stuck".
          console.warn(
            `Lobby websocket closed unexpectedly (code ${event.code}). Retrying in ${RECONNECT_DELAY_MS}ms…`,
          );
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

  // REST fallback: keep polling for fresh state as long as we don't have
  // a live connection, so the UI never gets permanently stuck on a stale
  // snapshot even if the websocket can't connect at all.
  useEffect(() => {
    if (!playerId || status === "open") return;

    let cancelled = false;

    const poll = () => {
      getLobby(lobbyId)
        .then((fresh) => {
          if (!cancelled) setLobby(normaliseLobby(fresh));
        })
        .catch(() => {
          // Ignore — the websocket retry loop already surfaces
          // connectivity issues, and this will just try again shortly.
        });
    };

    poll();
    const interval = setInterval(poll, FALLBACK_POLL_MS);

    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [lobbyId, playerId, status]);

  const send = useCallback((message: ClientMessage) => {
    const socket = socketRef.current;

    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(message));
    }
  }, []);

  return { lobby, status, lastError, send };
}
