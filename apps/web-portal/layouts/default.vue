<script setup lang="ts">
const config = useRuntimeConfig()

// Use useHead to load the Google Sign-In script
useHead({
  script: [
    { 
      src: 'https://accounts.google.com/gsi/client', 
      async: true, 
      defer: true 
    }
  ]
})

// Define the callback function to handle successful login
const handleLoginSuccess = async (response: any) => {
  // console.log('Google Login Success! Credential:', response.credential)
  
  try {
    // Send the credential to the identity API
    const data = await $fetch<any>(`${config.public.identityApiUrl}/auth/google`, {
      method: 'POST',
      body: { 
        credential: response.credential 
      }
    })
  
    if (data && data.display_name) {
      const authToken = useCookie('auth_token', {
        maxAge: 60 * 60 * 24 * 7, // 1 week
        sameSite: 'lax',
        secure: process.env.NODE_ENV === 'production'
      })
      authToken.value = data.display_name
      
      // Refresh the user state and navigate to the home page
      await navigateTo('/', { replace: true })
    }
  } catch (err) {
    console.error('Error connecting to Identity API:', err)
  }
}

// Initialize and render the button once the component is mounted
onMounted(() => {
  const initGoogleLogin = () => {
    if (typeof window !== 'undefined' && (window as any).google) {
      const google = (window as any).google;
      google.accounts.id.initialize({
        client_id: config.public.googleClientId,
        callback: handleLoginSuccess,
      });
      google.accounts.id.renderButton(
        document.getElementById('google-login-btn'),
        { theme: 'outline', size: 'large', type: 'standard' }
      );
    } else {
      // If the script isn't loaded yet, try again in a bit
      setTimeout(initGoogleLogin, 100);
    }
  }
  initGoogleLogin();
})
</script>

<template>
  <div class="min-h-screen flex flex-col bg-slate-900 font-sans antialiased text-slate-300">
    <AppHeader />
    <main class="flex-grow">
      <slot />
    </main>
  </div>
</template>
