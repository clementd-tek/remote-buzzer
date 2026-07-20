import { useEffect, useState } from "react";
import "./Buzzer.css";

export type BuzzerVisualState = "idle" | "countdown" | "go" | "win" | "lose";

interface BuzzerProps {
  state: BuzzerVisualState;
  label: string;
  onPress: () => void;
  /** Whole seconds remaining, shown large inside the buzzer face while state is "countdown". */
  countdownValue?: number | null;
}

export function Buzzer({ state, label, onPress, countdownValue }: BuzzerProps) {
  const [pressed, setPressed] = useState(false);
  const clickable = state === "go";

  // Re-arm the press animation any time the state changes, so a fresh
  // "go" always starts from a clean breathing state rather than a
  // leftover pressed-in look from the previous round.
  useEffect(() => {
    setPressed(false);
  }, [state]);

  function handleClick() {
    if (!clickable) return;
    setPressed(true);
    onPress();
  }

  return (
    <div className={`buzzer buzzer--${state}`}>
      <button
        type="button"
        className={`buzzer__button ${pressed ? "buzzer__button--pressed" : ""}`}
        onClick={handleClick}
        disabled={!clickable}
        aria-label={label}
      >
        <span className="buzzer__ring" />
        <span className="buzzer__face">
          {state === "countdown" && countdownValue != null ? (
            <span className="buzzer__countdown" key={countdownValue}>
              {countdownValue > 0 ? countdownValue : "GO"}
            </span>
          ) : (
            <span className="buzzer__icon" aria-hidden="true" />
          )}
        </span>
      </button>
      <p className="buzzer__label">{label}</p>
    </div>
  );
}
