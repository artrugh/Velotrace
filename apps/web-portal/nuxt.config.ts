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
        interval: 1000,
      },
    },
  },
  devServer: {
    host: "0.0.0.0",
    port: parseInt(process.env.NUXT_PORT || "3000", 10),
  },
  runtimeConfig: {
    identityApiUrl: process.env.IDENTITY_API_URL || "",
    bikesApiUrl: process.env.BIKES_API_URL || "",
    authCookieName:
      process.env.NODE_ENV === "production"
        ? "__Host-auth-token"
        : "auth-token",
    public: {
      googleClientId: process.env.GOOGLE_CLIENT_ID || "",
    },
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
});
