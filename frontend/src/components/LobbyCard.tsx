import { Link } from "react-router-dom";
import type { Lobby } from "../types/lobby";
import { StatusDot } from "./StatusDot";
import "./LobbyCard.css";

export function LobbyCard({ lobby }: { lobby: Lobby }) {
  return (
    <Link to={`/lobby/${lobby.id}`} className="lobby-card">
      <div className="lobby-card__main">
        <span className="lobby-card__name">{lobby.name}</span>
        <StatusDot state={lobby.state} />
      </div>
      <span className="lobby-card__count mono">
        {lobby.playerCount} {lobby.playerCount > 1 ? "joueurs" : "joueur"}
      </span>
    </Link>
  );
}
