export default defineEventHandler((event) => {
  deleteCookie(event, "auth-token", {
    path: "/",
  });
  return { success: true };
});
