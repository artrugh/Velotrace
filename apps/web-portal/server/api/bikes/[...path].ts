export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig();
  const token = getCookie(event, "auth-token");
  const targetUrl = event.path.replace(
    /^\/api\/bikes/,
    config.bikesApiUrl as string,
  );

  return proxyRequest(event, targetUrl, {
    headers: {
      Authorization: token ? `Bearer ${token}` : undefined,
      "X-Request-Id": crypto.randomUUID(),
    },
  });
});
