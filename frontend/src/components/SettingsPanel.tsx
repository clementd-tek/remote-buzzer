import { useState } from "react";
import { MAX_COUNTDOWN_SECONDS, MAX_POINTS_PER_ROUND, type LobbySettings } from "../types/lobby";
import "./SettingsPanel.css";

interface SettingsPanelProps {
  settings: LobbySettings;
  disabled: boolean; // true when a round is in flight
  onChange: (update: Partial<LobbySettings>) => void;
}

export function SettingsPanel({ settings, disabled, onChange }: SettingsPanelProps) {
  const [open, setOpen] = useState(false);

  return (
    <div className="settings-panel">
      <button
        type="button"
        className="settings-panel__toggle"
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
      >
        <span>⚙ Règles du jeu</span>
        <span className="settings-panel__caret">{open ? "▲" : "▼"}</span>
      </button>

      {open && (
        <div className="settings-panel__body">
          {disabled && (
            <p className="settings-panel__locked">
              Impossible de modifier les règles pendant une manche en cours.
            </p>
          )}

          <label className="settings-panel__field">
            <div className="settings-panel__field-header">
              <span>Points par manche</span>
              <span className="settings-panel__value mono">{settings.pointsPerRound} pt{settings.pointsPerRound > 1 ? "s" : ""}</span>
            </div>
            <input
              type="range"
              min={0}
              max={MAX_POINTS_PER_ROUND}
              value={settings.pointsPerRound}
              disabled={disabled}
              onChange={(e) => onChange({ pointsPerRound: Number(e.target.value) })}
            />
            <div className="settings-panel__range-labels">
              <span>0 (sans score)</span>
              <span>{MAX_POINTS_PER_ROUND}</span>
            </div>
          </label>

          <label className="settings-panel__field">
            <div className="settings-panel__field-header">
              <span>Compte à rebours</span>
              <span className="settings-panel__value mono">
                {settings.countdownSeconds === 0 ? "Instantané" : `${settings.countdownSeconds}s`}
              </span>
            </div>
            <input
              type="range"
              min={0}
              max={MAX_COUNTDOWN_SECONDS}
              value={settings.countdownSeconds}
              disabled={disabled}
              onChange={(e) => onChange({ countdownSeconds: Number(e.target.value) })}
            />
            <div className="settings-panel__range-labels">
              <span>Instantané</span>
              <span>{MAX_COUNTDOWN_SECONDS}s</span>
            </div>
          </label>
        </div>
      )}
    </div>
  );
}
