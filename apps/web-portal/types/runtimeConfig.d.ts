// /src/apps/web-portal/types/runtimeConfig.d.ts
import { RuntimeConfig } from "@nuxt/schema";

declare module "@nuxt/schema" {
  interface RuntimeConfig {
    identityApiUrl: string;
    bikesApiUrl: string;
    authCookieName: string;
  }
  interface PublicRuntimeConfig {
    googleClientId: string;
  }
}

// This is the "magic" line for Nitro/Server handlers
declare module "nitropack" {
  interface NitroRuntimeConfig {
    identityApiUrl: string;
    bikesApiUrl: string;
    authCookieName: string;
  }
}

export {};
