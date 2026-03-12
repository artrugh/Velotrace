export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss'],
  typescript: {
    strict: true,
    typeCheck: true
  },
  devServer: {
    host: '0.0.0.0',
    port: 3000
  },
  nitro: {
    preset: 'vercel',
    output: {
      dir: process.env.NITRO_OUTPUT_DIR || '../../.vercel/output'
    }
  }
})
