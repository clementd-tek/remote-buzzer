import { useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ApiError, getLobby, joinLobby } from "../api/client";
import { getRememberedHostId, getRememberedPlayer, newId, rememberPlayer } from "../api/identity";
import { isMuted, playGo, playLose, playRoundEnd, playTick, playWin, setMuted } from "../audio/sounds";
import { Buzzer, type BuzzerVisualState } from "../components/Buzzer";
import { HostControls } from "../components/HostControls";
import { JoinForm } from "../components/JoinForm";
import { PlayerRoster } from "../components/PlayerRoster";
import { Scoreboard } from "../components/Scoreboard";
import { StatusDot } from "../components/StatusDot";
import { COUNTDOWN_SECONDS } from "../constants";
import { useCountdown } from "../hooks/useCountdown";
import { useLobbySocket } from "../hooks/useLobbySocket";
import type { Lobby, LobbyState } from "../types/lobby";
import "./LobbyPage.css";

type Identity = { role: "host"; id: string } | { role: "player"; id: string };

export function LobbyPage() {
  const { id: lobbyId } = useParams<{ id: string }>();
  const [initialLobby, setInitialLobby] = useState<Lobby | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [identity, setIdentity] = useState<Identity | null>(null);
  const [checkingIdentity, setCheckingIdentity] = useState(true);

  useEffect(() => {
    if (!lobbyId) return;

    let cancelled = false;

    getLobby(lobbyId)
      .then((lobby) => {
        if (cancelled) return;

        setInitialLobby(lobby);

        const hostId = getRememberedHostId(lobbyId);
        if (hostId && hostId === lobby.hostId) {
          setIdentity({ role: "host", id: hostId });
          setCheckingIdentity(false);
          return;
        }

        const player = getRememberedPlayer(lobbyId);
        if (player) {
          setIdentity({ role: "player", id: player.id });
        }

        setCheckingIdentity(false);
      })
      .catch((err) => {
        if (cancelled) return;

        if (err instanceof ApiError && err.status === 404) {
          setNotFound(true);
        }

        setCheckingIdentity(false);
      });

    return () => {
      cancelled = true;
    };
  }, [lobbyId]);

  if (!lobbyId) return null;

  if (notFound) {
    return (
      <div className="lobby-page__center container">
        <h1>Lobby introuvable</h1>
        <p>Ce lien ne correspond à aucun lobby actif.</p>
        <Link to="/" className="lobby-page__back">
          Retour à l'accueil
        </Link>
      </div>
    );
  }

  if (checkingIdentity || !initialLobby) {
    return (
      <div className="lobby-page__center container">
        <p>Chargement du lobby…</p>
      </div>
    );
  }

  if (!identity) {
    return (
      <div className="container">
        <JoinForm
          lobbyName={initialLobby.name}
          onJoin={async (name) => {
            const playerId = newId();
            await joinLobby(lobbyId, { id: playerId, name });
            rememberPlayer(lobbyId, { id: playerId, name });
            setIdentity({ role: "player", id: playerId });
          }}
        />
      </div>
    );
  }

  return <ConnectedLobby lobbyId={lobbyId} identity={identity} fallback={initialLobby} />;
}

function buzzerState(lobby: Lobby, playerId: string, connected: boolean): BuzzerVisualState {
  if (lobby.state === "countdown") return "countdown";
  if (lobby.state === "open" && connected) return "go";

  if (lobby.state === "locked") {
    return lobby.winner?.playerId === playerId ? "win" : "lose";
  }

  return "idle";
}

function buzzerLabel(lobby: Lobby, state: BuzzerVisualState, connected: boolean): string {
  switch (state) {
    case "countdown":
      return "Prépare-toi !";
    case "go":
      return "Buzz !";
    case "win":
      return "Gagné !";
    case "lose":
      return "Trop lent";
    default:
      if (lobby.state === "open" && !connected) return "Reconnexion…";
      return lobby.state === "ready" ? "La manche va commencer…" : "En attente du lancement";
  }
}

/**
 * Plays a sound whenever the lobby's state actually transitions (not on
 * every broadcast — plenty of those don't change state, e.g. someone
 * else joining while we're still waiting). Skips the very first snapshot
 * so refreshing mid-round doesn't fire a sound for a transition that
 * didn't just happen.
 */
function useRoundSounds(lobby: Lobby, role: "host" | "player", myId: string) {
  const prevState = useRef<LobbyState | null>(null);

  useEffect(() => {
    const previous = prevState.current;
    prevState.current = lobby.state;

    if (previous === null || previous === lobby.state) return;

    if (lobby.state === "open") {
      playGo();
      return;
    }

    if (lobby.state === "locked") {
      if (role === "host") {
        playRoundEnd();
      } else if (lobby.winner?.playerId === myId) {
        playWin();
      } else {
        playLose();
      }
    }
  }, [lobby.state, lobby.winner, role, myId]);
}

function ConnectedLobby({
  lobbyId,
  identity,
  fallback,
}: {
  lobbyId: string;
  identity: Identity;
  fallback: Lobby;
}) {
  const { lobby: liveLobby, status, lastError, send } = useLobbySocket(lobbyId, identity.id);
  const lobby = liveLobby ?? fallback;
  const inviteUrl = `${window.location.origin}/lobby/${lobbyId}`;
  const [muted, setMutedState] = useState(() => isMuted());

  const countdownValue = useCountdown(lobby.countdownEndsAt, (seconds) => {
    if (seconds > 0) playTick();
  });

  useRoundSounds(lobby, identity.role, identity.id);

  function toggleMute() {
    const next = !muted;
    setMuted(next);
    setMutedState(next);
  }

  return (
    <div className="lobby-page container">
      <header className="lobby-page__header">
        <div>
          <h1 className="lobby-page__title">{lobby.name}</h1>
          <div className="lobby-page__meta">
            <StatusDot state={lobby.state} />
            <span className="lobby-page__round mono">Manche {lobby.roundNumber}</span>
          </div>
        </div>

        <button
          type="button"
          className="lobby-page__mute"
          onClick={toggleMute}
          aria-label={muted ? "Activer le son" : "Couper le son"}
          title={muted ? "Activer le son" : "Couper le son"}
        >
          {muted ? "🔇" : "🔊"}
        </button>
      </header>

      {status !== "open" && (
        <div className={`lobby-page__connection lobby-page__connection--${status}`}>
          {status === "connecting" && "Connexion en temps réel…"}
          {status === "reconnecting" &&
            "Connexion en temps réel perdue — nouvelle tentative en cours. Les données ci-dessous peuvent être en retard."}
          {status === "closed" && "Déconnecté."}
        </div>
      )}

      {lastError && <p className="lobby-page__error">{lastError}</p>}

      {identity.role === "host" ? (
        <div className="lobby-page__grid">
          <HostControls
            lobby={lobby}
            inviteUrl={inviteUrl}
            connected={status === "open"}
            countdownValue={countdownValue}
            onReady={() => send({ type: "ready" })}
            onOpen={() => send({ type: "open", seconds: COUNTDOWN_SECONDS })}
            onNextRound={() => send({ type: "next_round" })}
          />
          <div className="lobby-page__roster">
            <h2 className="lobby-page__roster-title">Joueurs</h2>
            <PlayerRoster
              players={lobby.players}
              winner={lobby.winner}
              hostId={lobby.hostId}
              currentPlayerId={null}
            />
          </div>
        </div>
      ) : (
        <div className="lobby-page__grid lobby-page__grid--player">
          <Buzzer
            state={buzzerState(lobby, identity.id, status === "open")}
            label={buzzerLabel(lobby, buzzerState(lobby, identity.id, status === "open"), status === "open")}
            countdownValue={countdownValue}
            onPress={() => send({ type: "buzz", playerId: identity.id })}
          />
          <div className="lobby-page__roster">
            <h2 className="lobby-page__roster-title">Joueurs</h2>
            <PlayerRoster
              players={lobby.players}
              winner={lobby.winner}
              hostId={lobby.hostId}
              currentPlayerId={identity.id}
            />
            <Scoreboard scores={lobby.scores} currentPlayerId={identity.id} />
          </div>
        </div>
      )}
    </div>
  );
}
