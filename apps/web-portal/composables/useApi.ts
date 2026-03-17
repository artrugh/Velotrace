import createClient from "openapi-fetch";
import type { paths as IdentityPaths } from "@api-contract/identity-api/generated/identity";

export const useIdentityApi = () => {
  const config = useRuntimeConfig();

  const client = createClient<IdentityPaths>({
    baseUrl: config.public.identityApiUrl,
  });

  return client;
};
