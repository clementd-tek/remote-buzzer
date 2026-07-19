import { useState, type FormEvent } from "react";
import "./JoinForm.css";

interface JoinFormProps {
  lobbyName: string;
  onJoin: (name: string) => Promise<void>;
}

export function JoinForm({ lobbyName, onJoin }: JoinFormProps) {
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();

    if (!name.trim()) {
      setError("Entre un prénom pour rejoindre.");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      await onJoin(name.trim());
    } catch {
      setError("Impossible de rejoindre. La manche a peut-être déjà commencé.");
      setSubmitting(false);
    }
  }

  return (
    <div className="join-form">
      <span className="join-form__eyebrow">Tu as été invité</span>
      <h1 className="join-form__title">{lobbyName}</h1>
      <p className="join-form__hint">Entre ton prénom pour monter sur scène.</p>

      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Ton prénom"
          maxLength={60}
          autoFocus
          autoComplete="off"
        />

        {error && <p className="join-form__error">{error}</p>}

        <button type="submit" disabled={submitting}>
          {submitting ? "Connexion…" : "Rejoindre"}
        </button>
      </form>
    </div>
  );
}
