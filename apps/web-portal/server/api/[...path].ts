export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig();
  const path = getRouterParam(event, "path") || "";
  const targetUrl = `${config.bikesApiUrl}/${path}`;

  console.log(
    `[BFF Global Proxy] ${event.method} ${event.path} -> ${targetUrl}`,
  );

  const token = getCookie(event, config.authCookieName);

  try {
    return await proxyRequest(event, targetUrl, {
      headers: {
        Authorization: token ? `Bearer ${token}` : undefined,
        "X-Request-Id": crypto.randomUUID(),
      },
    });
  } catch (e: any) {
    console.error("[BFF Proxy Error]:", e.message);
    throw createError({
      statusCode: 502,
      statusMessage: "Backend unreachable",
    });
  }
});
