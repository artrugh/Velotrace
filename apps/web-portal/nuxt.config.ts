import { defineNuxtConfig } from "nuxt/config";

export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ["@nuxtjs/tailwindcss", "@nuxt/test-utils/module"],
  typescript: {
    strict: false,
    typeCheck: true,
  },
  devServer: {
    host: "0.0.0.0",
    port: 3000,
  },
  runtimeConfig: {
    public: {
      googleClientId: process.env.GOOGLE_CLIENT_ID,
      identityApiUrl: process.env.IDENTITY_API_URL,
    },
  },
  nitro: {
    preset: "vercel",
    output: {
      dir: process.env.NITRO_OUTPUT_DIR,
    },
  },
} as Parameters<typeof defineNuxtConfig>[0] & { nitro?: any });
// Workaround: NuxtConfig interface omits 'nitro' property from ConfigSchema.
// Using intersection type to re-add 'nitro' to bypass TypeScript error ts(2353).
