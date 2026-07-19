import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ApiError, getLobby, joinLobby } from "../api/client";
import { getRememberedHostId, getRememberedPlayer, newId, rememberPlayer } from "../api/identity";
import { Buzzer, type BuzzerVisualState } from "../components/Buzzer";
import { HostControls } from "../components/HostControls";
import { JoinForm } from "../components/JoinForm";
import { PlayerRoster } from "../components/PlayerRoster";
import { StatusDot } from "../components/StatusDot";
import { useLobbySocket } from "../hooks/useLobbySocket";
import type { Lobby } from "../types/lobby";
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

function buzzerState(lobby: Lobby, playerId: string): BuzzerVisualState {
  if (lobby.state === "open") return "go";

  if (lobby.state === "locked") {
    return lobby.winner?.playerId === playerId ? "win" : "lose";
  }

  return "idle";
}

function buzzerLabel(lobby: Lobby, state: BuzzerVisualState): string {
  switch (state) {
    case "go":
      return "Buzz !";
    case "win":
      return "Gagné !";
    case "lose":
      return "Trop lent";
    default:
      return lobby.state === "ready" ? "La manche va commencer…" : "En attente du lancement";
  }
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

  return (
    <div className="lobby-page container">
      <header className="lobby-page__header">
        <div>
          <h1 className="lobby-page__title">{lobby.name}</h1>
          <StatusDot state={lobby.state} />
        </div>

        {status === "reconnecting" && (
          <span className="lobby-page__reconnect">Connexion perdue, reconnexion…</span>
        )}
      </header>

      {lastError && <p className="lobby-page__error">{lastError}</p>}

      {identity.role === "host" ? (
        <div className="lobby-page__grid">
          <HostControls
            lobby={lobby}
            inviteUrl={inviteUrl}
            onReady={() => send({ type: "ready" })}
            onOpen={() => send({ type: "open" })}
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
            state={buzzerState(lobby, identity.id)}
            label={buzzerLabel(lobby, buzzerState(lobby, identity.id))}
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
          </div>
        </div>
      )}
    </div>
  );
}
