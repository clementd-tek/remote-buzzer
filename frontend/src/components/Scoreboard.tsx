import type { Score } from "../types/lobby";
import "./Scoreboard.css";

interface ScoreboardProps {
  scores: Score[];
  currentPlayerId?: string | null;
}

export function Scoreboard({ scores, currentPlayerId }: ScoreboardProps) {
  if (scores.length === 0) {
    return null;
  }

  return (
    <div className="scoreboard">
      <h3 className="scoreboard__title">Classement</h3>
      <ol className="scoreboard__list">
        {scores.map((score, index) => (
          <li
            key={score.playerId}
            className={`scoreboard__row ${index === 0 ? "scoreboard__row--leader" : ""}`}
          >
            <span className="scoreboard__rank">{index + 1}</span>
            <span className="scoreboard__name">
              {score.name || "?"}
              {score.playerId === currentPlayerId && <span className="scoreboard__you">toi</span>}
            </span>
            <span className="scoreboard__points mono">
              {score.points} {score.points > 1 ? "pts" : "pt"}
            </span>
          </li>
        ))}
      </ol>
    </div>
  );
}
