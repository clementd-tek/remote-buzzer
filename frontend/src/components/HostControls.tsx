import { useState } from "react";
import type { Lobby, LobbySettings } from "../types/lobby";
import { Scoreboard } from "./Scoreboard";
import { SettingsPanel } from "./SettingsPanel";
import "./HostControls.css";

interface HostControlsProps {
  lobby: Lobby;
  inviteUrl: string;
  connected: boolean;
  countdownValue: number | null;
  onReady: () => void;
  onOpen: () => void;
  onNextRound: () => void;
  onSettingsChange: (update: Partial<LobbySettings>) => void;
}

export function HostControls({
  lobby,
  inviteUrl,
  connected,
  countdownValue,
  onReady,
  onOpen,
  onNextRound,
  onSettingsChange,
}: HostControlsProps) {
  const [copied, setCopied] = useState(false);

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(inviteUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard access can fail; link is still visible for manual copy.
    }
  }

  const winnerName = lobby.winner
    ? lobby.players.find((p) => p.id === lobby.winner!.playerId)?.name
    : undefined;

  // Settings can only be changed while waiting or between rounds.
  const settingsLocked = lobby.state === "countdown" || lobby.state === "open" || lobby.state === "locked";

  const countdownSec = lobby.settings?.countdownSeconds ?? 0;
  const countdownLabel = countdownSec === 0 ? "Instantané" : `${countdownSec}s`;

  return (
    <div className="host-controls">
      <div className="host-controls__invite">
        <span className="host-controls__invite-label">Lien à partager</span>
        <div className="host-controls__invite-row">
          <code className="mono">{inviteUrl}</code>
          <button type="button" onClick={copyLink}>
            {copied ? "Copié !" : "Copier"}
          </button>
        </div>
      </div>

      <SettingsPanel
        settings={lobby.settings ?? { pointsPerRound: 1, countdownSeconds: 3 }}
        disabled={settingsLocked}
        onChange={onSettingsChange}
      />

      <div className="host-controls__actions">
        <span className="host-controls__round mono">Manche {lobby.roundNumber}</span>

        {lobby.state === "waiting" && (
          <>
            <p>
              {lobby.playerCount === 0
                ? "En attente du premier joueur…"
                : `${lobby.playerCount} joueur${lobby.playerCount > 1 ? "s" : ""} connecté${lobby.playerCount > 1 ? "s" : ""}.`}
            </p>
            <button
              type="button"
              className="host-controls__primary"
              onClick={onReady}
              disabled={lobby.playerCount === 0 || !connected}
            >
              Verrouiller les inscriptions
            </button>
          </>
        )}

        {lobby.state === "ready" && (
          <>
            <p>Inscriptions verrouillées. Prêt quand tu veux.</p>
            <button
              type="button"
              className="host-controls__primary host-controls__primary--go"
              onClick={onOpen}
              disabled={!connected}
            >
              Lancer le buzzer ({countdownLabel})
            </button>
          </>
        )}

        {lobby.state === "countdown" && (
          <p className="host-controls__countdown">
            Compte à rebours : <strong>{countdownValue ?? countdownSec}</strong>
          </p>
        )}

        {lobby.state === "open" && <p className="host-controls__live">Le buzzer est en direct.</p>}

        {lobby.state === "locked" && (
          <>
            <p className="host-controls__result">
              {winnerName ? (
                <>
                  🏆 <strong>{winnerName}</strong> a buzzé en premier.
                </>
              ) : (
                "La manche est terminée."
              )}
            </p>
            <button
              type="button"
              className="host-controls__primary host-controls__primary--go"
              onClick={onNextRound}
              disabled={!connected}
            >
              Manche suivante
            </button>
          </>
        )}
      </div>

      <Scoreboard scores={lobby.scores} />
    </div>
  );
}
