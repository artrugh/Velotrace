import createClient from "openapi-fetch";
import type {
  paths as IdentityPaths,
  components as IdentityComponents,
} from "@api-contract/.generated/identity";

export default defineEventHandler(async (event) => {
  const body =
    await readBody<IdentityComponents["schemas"]["handler.AuthGoogleRequest"]>(
      event,
    );
  const config = useRuntimeConfig();

  const identityClient = createClient<IdentityPaths>({
    baseUrl: config.identityApiUrl as string,
  });

  const { data, error } = await identityClient.POST("/auth/google", {
    body,
  });

  if (error) {
    throw createError({
      statusCode: 401,
      statusMessage: "Authentication failed",
      data: error,
    });
  }

  if (data && data.token) {
    // Set the cookie for the entire domain, shared across client and server
    setCookie(event, "auth-token", data.token, {
      maxAge: 60 * 60 * 24 * 7, // 1 week
      sameSite: "lax",
      path: "/",
      secure: process.env.NODE_ENV === "production",
      httpOnly: true,
    });
  }

  return { success: true };
});
