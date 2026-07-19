import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

// Without this, a component left mounted between tests (e.g. one holding
// an open websocket with a reconnect timer) keeps running in the
// background and can throw once the test environment starts tearing
// down, surfacing as a confusing "unhandled error" on an unrelated test.
afterEach(() => {
  cleanup();
});

// Known Node/undici + jsdom interop quirk, unrelated to application
// code: undici's WebSocket implementation dispatches connection events
// using Node's native Event class, but jsdom's environment replaces the
// global Event constructor with its own — so if a socket's handshake
// resolves asynchronously after a test has already unmounted its
// component (a real race in a websocket-with-reconnect hook under real
// network timing), undici's internal dispatch throws an "instanceof"
// mismatch. This never happens in an actual browser, which only has one
// Event class. Only this exact, identified error is swallowed; anything
// else still fails tests normally.
process.on("uncaughtException", (err) => {
  const isKnownUndiciJsdomQuirk =
    err instanceof TypeError && err.message.includes('must be an instance of Event');

  if (!isKnownUndiciJsdomQuirk) {
    throw err;
  }
});

