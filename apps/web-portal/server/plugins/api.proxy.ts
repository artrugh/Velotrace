export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook("request", (event) => {
    // Minimal logging for incoming requests
    console.log(`[Nitro] Incoming request: ${event.method} ${event.path}`);
  });
});
