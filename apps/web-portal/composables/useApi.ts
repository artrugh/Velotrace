import createClient from "openapi-fetch";
import type { paths as IdentityPaths } from "@api-contract/.generated/identity";
import type { paths as BikesPaths } from "@api-contract/.generated/bikes";
export const useIdentityApi = () => {
  const config = useRuntimeConfig();

  const client = createClient<IdentityPaths>({
    baseUrl: config.public.identityApiUrl,
  });

  return client;
};

export const useBikesApi = () => {
  const config = useRuntimeConfig();

  const client = createClient<BikesPaths>({
    baseUrl: config.public.bikesApiUrl,
  });

  return client;
};
