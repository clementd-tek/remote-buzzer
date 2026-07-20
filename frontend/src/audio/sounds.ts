// All sounds here are synthesized on the fly with oscillators rather than
// loaded from audio files — no assets to bundle, no licensing to worry
// about, and it keeps this feature purely client-side.

const MUTE_KEY = "buzzer:muted";

let ctx: AudioContext | null = null;

function getContext(): AudioContext | null {
  if (typeof window === "undefined") return null;

  const AudioContextClass = window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;

  if (!AudioContextClass) return null;

  if (!ctx) {
    ctx = new AudioContextClass();
  }

  // Browsers suspend the context until a user gesture; by the time any
  // sound plays here the person has already typed/clicked something, so
  // this just resumes a context that's allowed to run.
  if (ctx.state === "suspended") {
    void ctx.resume();
  }

  return ctx;
}

export function isMuted(): boolean {
  if (typeof window === "undefined") return true;
  return window.localStorage.getItem(MUTE_KEY) === "true";
}

export function setMuted(muted: boolean): void {
  window.localStorage.setItem(MUTE_KEY, String(muted));
}

interface ToneOptions {
  frequency: number;
  startOffset: number;
  duration: number;
  type?: OscillatorType;
  peakGain?: number;
}

function scheduleTone(audioCtx: AudioContext, { frequency, startOffset, duration, type = "sine", peakGain = 0.2 }: ToneOptions) {
  const oscillator = audioCtx.createOscillator();
  const gain = audioCtx.createGain();

  oscillator.type = type;
  oscillator.frequency.value = frequency;

  const startTime = audioCtx.currentTime + startOffset;
  const endTime = startTime + duration;

  // Quick attack, exponential-ish decay — avoids the click of a hard
  // on/off and sounds far less harsh than a flat volume.
  gain.gain.setValueAtTime(0, startTime);
  gain.gain.linearRampToValueAtTime(peakGain, startTime + Math.min(0.015, duration / 4));
  gain.gain.exponentialRampToValueAtTime(0.0001, endTime);

  oscillator.connect(gain);
  gain.connect(audioCtx.destination);

  oscillator.start(startTime);
  oscillator.stop(endTime + 0.02);
  oscillator.onended = () => {
    oscillator.disconnect();
    gain.disconnect();
  };
}

function play(tones: ToneOptions[]) {
  if (isMuted()) return;

  const audioCtx = getContext();
  if (!audioCtx) return;

  for (const tone of tones) {
    scheduleTone(audioCtx, tone);
  }
}

/** Short blip for each countdown second (3, 2, 1...). */
export function playTick(): void {
  play([{ frequency: 880, startOffset: 0, duration: 0.09, type: "sine", peakGain: 0.18 }]);
}

/** Two-note "starting bell" for when the buzzer goes live. */
export function playGo(): void {
  play([
    { frequency: 523.25, startOffset: 0, duration: 0.12, type: "triangle", peakGain: 0.25 },
    { frequency: 784.0, startOffset: 0.1, duration: 0.18, type: "triangle", peakGain: 0.28 },
  ]);
}

/** Triumphant ascending arpeggio for winning a round. */
export function playWin(): void {
  play([
    { frequency: 523.25, startOffset: 0, duration: 0.13, type: "triangle", peakGain: 0.22 },
    { frequency: 659.25, startOffset: 0.11, duration: 0.13, type: "triangle", peakGain: 0.22 },
    { frequency: 784.0, startOffset: 0.22, duration: 0.13, type: "triangle", peakGain: 0.22 },
    { frequency: 1046.5, startOffset: 0.33, duration: 0.28, type: "triangle", peakGain: 0.25 },
  ]);
}

/** Short descending buzz for losing a round. */
export function playLose(): void {
  play([
    { frequency: 220, startOffset: 0, duration: 0.16, type: "sawtooth", peakGain: 0.14 },
    { frequency: 146.83, startOffset: 0.13, duration: 0.24, type: "sawtooth", peakGain: 0.16 },
  ]);
}

/** Neutral chime for the host when a round ends (no personal win/lose framing). */
export function playRoundEnd(): void {
  play([
    { frequency: 659.25, startOffset: 0, duration: 0.12, type: "sine", peakGain: 0.2 },
    { frequency: 523.25, startOffset: 0.1, duration: 0.2, type: "sine", peakGain: 0.18 },
  ]);
}
