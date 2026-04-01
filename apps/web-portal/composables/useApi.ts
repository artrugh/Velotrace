import createClient from "openapi-fetch";
import type { paths as BikesPaths } from "@api-contract/.generated/bikes";

export const useBikesApi = () => {
  const client = createClient<BikesPaths>({
    baseUrl: "/api/bikes", // This will hit server/api/bikes/[...path].ts
  });

  return client;
};
