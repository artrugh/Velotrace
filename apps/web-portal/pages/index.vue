<template>
  <div
    class="flex items-center justify-center py-20 px-4 text-center max-w-2xl mx-auto"
  >
    <div>
      <h2 class="text-5xl font-extrabold text-white mb-6 tracking-tight">
        Welcome to <span class="text-blue-500">VeloTrace</span>
      </h2>
      <p class="text-slate-400 text-xl leading-relaxed mb-10">
        The
        <span class="text-slate-200 font-medium italic">Silent Sentry</span>
        for your bicycle property. A secure, registry-first bicycle ownership
        platform.
      </p>

      <div class="flex gap-4 justify-center">
        <NuxtLink
          to="/bikes/register"
          class="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition-all shadow-lg shadow-blue-900/20"
        >
          Register a Bike
        </NuxtLink>
        <button
          class="px-6 py-3 bg-slate-800 hover:bg-slate-700 text-slate-300 font-semibold rounded-lg transition-all border border-slate-700"
        >
          Browse Marketplace
        </button>
      </div>
    </div>
  </div>
</template>
