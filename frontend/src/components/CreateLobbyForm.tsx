import { useState, type FormEvent } from "react";
import "./CreateLobbyForm.css";

interface CreateLobbyFormProps {
  onCreate: (name: string, isPublic: boolean) => Promise<void>;
}

export function CreateLobbyForm({ onCreate }: CreateLobbyFormProps) {
  const [name, setName] = useState("");
  const [isPublic, setIsPublic] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();

    if (!name.trim()) {
      setError("Donne un nom à ton lobby.");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      await onCreate(name.trim(), isPublic);
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

      {error && <p className="create-form__error">{error}</p>}

      <button type="submit" className="create-form__submit" disabled={submitting}>
        {submitting ? "Création…" : "Créer le lobby"}
      </button>
    </form>
  );
}
