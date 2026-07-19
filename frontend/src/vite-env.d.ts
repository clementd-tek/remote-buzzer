/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Override the API/websocket base URL. Leave unset to use relative
   * paths (proxied by Vite in dev, by nginx in production). */
  readonly VITE_API_BASE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
