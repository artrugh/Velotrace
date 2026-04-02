export default defineEventHandler(async (event) => {
  console.error('------- PROXY TRIGGERED BY SSR -------'); 

  const config = useRuntimeConfig();
  const subPath = getRouterParam(event, 'path') || "";
  const cleanSubPath = subPath ? (subPath.startsWith('/') ? subPath : `/${subPath}`) : "";  
  const targetUrl = `${config.bikesApiUrl}/bikes${cleanSubPath}`;
  console.log(`[BFF Proxy] Target: ${targetUrl}`);

  const token = getCookie(event, config.authCookieName);

  try {
    return await proxyRequest(event, targetUrl, {
      headers: {
        Authorization: token ? `Bearer ${token}` : undefined,
        "X-Request-Id": crypto.randomUUID(),
      },
    });
  } catch (e) {
    console.error('[BFF Proxy Error]:', e);
    throw createError({
      statusCode: 502,
      statusMessage: 'Bad Gateway - Failed to reach Bike API',
    });
  }
});