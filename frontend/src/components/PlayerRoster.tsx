import type { Player, Winner } from "../types/lobby";
import "./PlayerRoster.css";

interface PlayerRosterProps {
  players: Player[];
  winner?: Winner;
  currentPlayerId?: string | null;
  hostId: string;
}

export function PlayerRoster({ players, winner, currentPlayerId, hostId }: PlayerRosterProps) {
  if (players.length === 0) {
    return <p className="roster-empty">Personne n'a encore rejoint. Partage le lien !</p>;
  }

  return (
    <ul className="roster">
      {players.map((player) => {
        const isWinner = winner?.playerId === player.id;
        const isYou = player.id === currentPlayerId;
        const isLoser = Boolean(winner) && !isWinner;

        return (
          <li
            key={player.id}
            className={`roster__tag ${isWinner ? "roster__tag--winner" : ""} ${
              isLoser ? "roster__tag--loser" : ""
            }`}
          >
            <span className="roster__name">{player.name}</span>
            {isYou && <span className="roster__you">toi</span>}
            {player.id === hostId && <span className="roster__host">hôte</span>}
          </li>
        );
      })}
    </ul>
  );
}
