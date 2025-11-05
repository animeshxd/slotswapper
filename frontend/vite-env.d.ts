/// <reference types="vite/client" />
/** biome-ignore-all lint/suspicious/noEmptyInterface: ignore */
/** biome-ignore-all lint/correctness/noUnusedVariables: ingore */

interface ViteTypeOptions {
}

interface ImportMetaEnv {
  readonly VITE_HTTP_SERVER_URL: string
  // more env variables...
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}