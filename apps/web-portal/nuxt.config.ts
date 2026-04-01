import { defineNuxtConfig } from "nuxt/config";

export default defineNuxtConfig({
  devtools: { enabled: true },
  experimental: { appManifest: process.env.NODE_ENV !== "development" },
  modules: ["@nuxtjs/tailwindcss", "@nuxt/test-utils/module"],
  typescript: {
    strict: false,
    typeCheck: true,
  },
  vite: {
    server: {
      watch: {
        usePolling: true,
        interval: 1000, // Check for changes every 1 second
      },
    },
  },
  devServer: {
    host: "0.0.0.0",
    port: process.env.NUXT_PORT || 3000,
  },
  runtimeConfig: {
    public: {
      googleClientId: process.env.GOOGLE_CLIENT_ID,
    },
    identityApiUrl: process.env.IDENTITY_API_URL,
    bikesApiUrl: process.env.BIKES_API_URL,
  },
  routeRules: {
    "/": {
      headers: {
        "Cross-Origin-Opener-Policy": "same-origin-allow-popups",
      },
    },
  },
  nitro: {
    preset: "vercel",
  },
} as Parameters<typeof defineNuxtConfig>[0] & { nitro?: any });
// Workaround: NuxtConfig interface omits 'nitro' property from ConfigSchema.
// Using intersection type to re-add 'nitro' to bypass TypeScript error ts(2353).
