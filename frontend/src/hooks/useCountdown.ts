import { useEffect, useRef, useState } from "react";

/**
 * Ticks down to an absolute end time (an ISO timestamp from the server),
 * rather than starting a fresh local timer — this way every connected
 * client shows the same countdown regardless of when its own message
 * arrived or how far its clock has drifted.
 *
 * Returns the whole seconds remaining (0 once it's elapsed), or null if
 * there's no countdown running. Calls onSecondChange each time the
 * displayed integer changes, so callers can hang a tick sound off it
 * without re-implementing the "did the number just change" check.
 */
export function useCountdown(endsAt: string | undefined, onSecondChange?: (seconds: number) => void): number | null {
  const [remaining, setRemaining] = useState<number | null>(null);
  const lastReported = useRef<number | null>(null);
  const onSecondChangeRef = useRef(onSecondChange);
  onSecondChangeRef.current = onSecondChange;

  useEffect(() => {
    if (!endsAt) {
      setRemaining(null);
      lastReported.current = null;
      return;
    }

    const endTime = new Date(endsAt).getTime();

    const tick = () => {
      const secondsLeft = Math.max(0, Math.ceil((endTime - Date.now()) / 1000));
      setRemaining(secondsLeft);

      if (lastReported.current !== secondsLeft) {
        lastReported.current = secondsLeft;
        onSecondChangeRef.current?.(secondsLeft);
      }
    };

    tick();
    const interval = setInterval(tick, 100);

    return () => clearInterval(interval);
  }, [endsAt]);

  return remaining;
}
