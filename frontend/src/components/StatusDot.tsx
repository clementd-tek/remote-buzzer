import type { LobbyState } from "../types/lobby";
import "./StatusDot.css";

const LABELS: Record<LobbyState, string> = {
  waiting: "En attente",
  ready: "Prêt à démarrer",
  open: "En direct",
  locked: "Terminé",
};

export function StatusDot({ state }: { state: LobbyState }) {
  return (
    <span className={`status-dot status-dot--${state}`}>
      <span className="status-dot__mark" />
      {LABELS[state]}
    </span>
  );
}
