import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // Fail loudly if 5173 is taken instead of silently picking another
    // port — a silent port switch is exactly what breaks the backend's
    // websocket origin check in confusing, hard-to-notice ways.
    strictPort: true,
    proxy: {
      // Forwards both REST calls and the websocket upgrade to the Go
      // backend during local development, so the app can always call
      // relative /api/... paths (see src/api/client.ts).
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        ws: true,
      },
    },
  },
});
