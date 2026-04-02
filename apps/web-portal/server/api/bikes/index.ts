export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig();

  // Since this is index.ts, the target is always the base resource
  // Result: bikesApiUrl/bikes
  const targetUrl = `${config.bikesApiUrl}/bikes`;

  console.log(`[BFF Proxy INDEX] Forwarding to: ${targetUrl}`);

  const token = getCookie(event, config.authCookieName);

  return await proxyRequest(event, targetUrl, {
    headers: {
      Authorization: token ? `Bearer ${token}` : undefined,
      "X-Request-Id": crypto.randomUUID(),
    },
  });
});
