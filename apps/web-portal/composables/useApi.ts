import createClient from "openapi-fetch";
import type {
  paths as BikesPaths,
  components as BikesComponents,
} from "@api-contract/.generated/bikes";

export type Bike = BikesComponents["schemas"]["domain.Bike"];
export type BikeImage = NonNullable<Bike["images"]>[number];

export const useBikesApi = () => {
  // 1. Grab the cookies from the incoming browser request (SSR only)
  const headers = useRequestHeaders(["cookie"]);
  // 2. Get the full URL so the server knows where to find "/api/bikes"
  const host = process.server ? useRequestURL().origin : "";

  const client = createClient<BikesPaths>({
    baseUrl: `${host}/api`,
    headers: headers as Record<string, string>,
  });

  return client;
};
