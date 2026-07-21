import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createLobby, listPublicLobbies } from "../api/client";
import { newId, rememberHost } from "../api/identity";
import { CreateLobbyForm } from "../components/CreateLobbyForm";
import { LobbyCard } from "../components/LobbyCard";
import type { Lobby, LobbySettings } from "../types/lobby";
import "./HomePage.css";

const POLL_INTERVAL_MS = 4000;

export function HomePage() {
  const navigate = useNavigate();
  const [lobbies, setLobbies] = useState<Lobby[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  const refresh = useCallback(async () => {
    try {
      const result = await listPublicLobbies();
      setLobbies(result);
      setLoadError(false);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [refresh]);

  async function handleCreate(name: string, isPublic: boolean, settings: LobbySettings) {
    const hostId = newId();
    const lobby = await createLobby({ name, hostId, public: isPublic, settings });
    rememberHost(lobby.id, hostId);
    navigate(`/lobby/${lobby.id}`);
  }

  return (
    <div className="home">
      <header className="home__hero container">
        <div className="home__hero-mark" aria-hidden="true">
          <span className="home__hero-dot" />
        </div>
        <span className="home__eyebrow">Buzzer en réseau, en temps réel</span>
        <h1 className="home__title">
          LE BUZZER<span className="home__title-dot">.</span>
        </h1>
        <p className="home__tagline">
          Crée un lobby, partage le lien, et laisse le serveur départager qui a appuyé en
          premier — comme à la télé.
        </p>
      </header>

      <main className="home__body container">
        <section className="home__create">
          <CreateLobbyForm onCreate={handleCreate} />
        </section>

        <section className="home__list">
          <div className="home__list-header">
            <h2>Lobbies en direct</h2>
            <span className="home__list-count mono">{lobbies.length}</span>
          </div>

          {loading && <p className="home__status">Chargement…</p>}

          {!loading && loadError && (
            <p className="home__status home__status--error">
              Impossible de contacter le serveur. Nouvelle tentative dans quelques secondes…
            </p>
          )}

          {!loading && !loadError && lobbies.length === 0 && (
            <p className="home__status">
              Aucun lobby public pour l'instant. Sois le premier à en créer un.
            </p>
          )}

          <div className="home__list-items">
            {lobbies.map((lobby) => (
              <LobbyCard key={lobby.id} lobby={lobby} />
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
