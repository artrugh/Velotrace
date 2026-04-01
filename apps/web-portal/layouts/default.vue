<script setup lang="ts">
useHead({
  script: [
    {
      src: "https://accounts.google.com/gsi/client",
      async: true,
      defer: true,
    },
  ],
});

const handleLoginSuccess = async (response: { credential: string }) => {
  try {
    const { data, error } = await useFetch("/api/login", {
      method: "POST",
      body: {
        credential: response.credential,
      },
    });

    if (error.value) {
      console.error("Login Error:", error.value);
      throw new Error("Login failed");
    }

    if (data.value) {
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
