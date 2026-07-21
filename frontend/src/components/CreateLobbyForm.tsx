import { useState, type FormEvent } from "react";
import { DEFAULT_SETTINGS, MAX_COUNTDOWN_SECONDS, MAX_POINTS_PER_ROUND, type LobbySettings } from "../types/lobby";
import "./CreateLobbyForm.css";

interface CreateLobbyFormProps {
  onCreate: (name: string, isPublic: boolean, settings: LobbySettings) => Promise<void>;
}

export function CreateLobbyForm({ onCreate }: CreateLobbyFormProps) {
  const [name, setName] = useState("");
  const [isPublic, setIsPublic] = useState(true);
  const [settings, setSettings] = useState<LobbySettings>(DEFAULT_SETTINGS);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showSettings, setShowSettings] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();

    if (!name.trim()) {
      setError("Donne un nom à ton lobby.");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      await onCreate(name.trim(), isPublic, settings);
    } catch {
      setError("Impossible de créer le lobby. Réessaie.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form className="create-form" onSubmit={handleSubmit}>
      <h2 className="create-form__title">Créer un lobby</h2>
      <p className="create-form__subtitle">Donne un nom, invite tes amis, lance la manche.</p>

      <label className="create-form__field">
        <span>Nom du lobby</span>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Questions pour un champion"
          maxLength={60}
          autoComplete="off"
        />
      </label>

      <label className="create-form__toggle">
        <input
          type="checkbox"
          checked={isPublic}
          onChange={(e) => setIsPublic(e.target.checked)}
        />
        <span>Afficher sur la page d'accueil</span>
      </label>

      {/* Collapsible advanced settings */}
      <div className="create-form__advanced">
        <button
          type="button"
          className="create-form__advanced-toggle"
          onClick={() => setShowSettings((s) => !s)}
          aria-expanded={showSettings}
        >
          <span>⚙ Règles du jeu</span>
          <span>{showSettings ? "▲" : "▼"}</span>
        </button>

        {showSettings && (
          <div className="create-form__settings-body">
            <label className="create-form__setting">
              <div className="create-form__setting-header">
                <span>Points par manche</span>
                <span className="mono create-form__setting-value">
                  {settings.pointsPerRound} pt{settings.pointsPerRound > 1 ? "s" : ""}
                </span>
              </div>
              <input
                type="range"
                min={0}
                max={MAX_POINTS_PER_ROUND}
                value={settings.pointsPerRound}
                onChange={(e) => setSettings((s) => ({ ...s, pointsPerRound: Number(e.target.value) }))}
              />
              <div className="create-form__range-labels">
                <span>0 (sans score)</span>
                <span>{MAX_POINTS_PER_ROUND}</span>
              </div>
            </label>

            <label className="create-form__setting">
              <div className="create-form__setting-header">
                <span>Compte à rebours</span>
                <span className="mono create-form__setting-value">
                  {settings.countdownSeconds === 0 ? "Instantané" : `${settings.countdownSeconds}s`}
                </span>
              </div>
              <input
                type="range"
                min={0}
                max={MAX_COUNTDOWN_SECONDS}
                value={settings.countdownSeconds}
                onChange={(e) => setSettings((s) => ({ ...s, countdownSeconds: Number(e.target.value) }))}
              />
              <div className="create-form__range-labels">
                <span>Instantané</span>
                <span>{MAX_COUNTDOWN_SECONDS}s</span>
              </div>
            </label>
          </div>
        )}
      </div>

      {error && <p className="create-form__error">{error}</p>}

      <button type="submit" className="create-form__submit" disabled={submitting}>
        {submitting ? "Création…" : "Créer le lobby"}
      </button>
    </form>
  );
}
