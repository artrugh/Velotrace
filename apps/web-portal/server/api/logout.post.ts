export default defineEventHandler((event) => {
  const config = useRuntimeConfig();
  deleteCookie(event, config.authCookieName, {
    path: "/",
  });
  return { success: true };
});
