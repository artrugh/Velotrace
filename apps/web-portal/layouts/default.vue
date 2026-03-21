<script setup lang="ts">
import { useIdentityApi } from "../composables/useApi";

useHead({
  script: [
    {
      src: "https://accounts.google.com/gsi/client",
      async: true,
      defer: true,
    },
  ],
});

const handleLoginSuccess = async (response: any) => {
  try {
    const { data, error } = await useIdentityApi().POST("/auth/google", {
      body: {
        credential: response.credential,
      },
    });

    if (error) {
      throw new Error(`API Error: ${JSON.stringify(error)}`);
    }

    if (data && data.token) {
      const authToken = useCookie("auth_token", {
        maxAge: 60 * 60 * 24 * 7, // 1 week
        sameSite: "lax",
        secure: process.env.NODE_ENV === "production",
        httpOnly: true
      });
      authToken.value = data.token;

      // Refresh the user state and navigate to the home page
      await navigateTo("/", { replace: true });
    }
  } catch (err) {
    console.error("Error connecting to Identity API:", err);
  }
};

onMounted(() => {
  const config = useRuntimeConfig();
  const initGoogleLogin = () => {
    if (typeof window !== "undefined" && (window as any).google) {
      const google = (window as any).google;
      google.accounts.id.initialize({
        client_id: config.public.googleClientId,
        callback: handleLoginSuccess,
      });
      google.accounts.id.renderButton(
        document.getElementById("google-login-btn"),
        { theme: "outline", size: "large", type: "standard" },
      );
    } else {
      setTimeout(initGoogleLogin, 100);
    }
  };
  initGoogleLogin();
});
</script>

<template>
  <div
    class="min-h-screen flex flex-col bg-slate-900 font-sans antialiased text-slate-300"
  >
    <AppHeader />
    <main class="flex-grow">
      <slot />
    </main>
  </div>
</template>
