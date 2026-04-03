<template>
  <div class="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
    <!-- Back Button -->
    <div class="mb-8">
      <NuxtLink
        to="/marketplace"
        class="text-slate-500 hover:text-white flex items-center gap-2 transition-colors group text-sm font-medium"
      >
        <svg
          class="w-4 h-4 transform group-hover:-translate-x-1 transition-transform"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 19l-7-7m0 0l7-7m-7 7h18"
          />
        </svg>
        BACK TO MARKETPLACE
      </NuxtLink>
    </div>

    <!-- Loading State -->
    <div v-if="pending" class="animate-pulse">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
        <div class="aspect-[4/3] bg-slate-800/50 rounded-2xl"></div>
        <div class="space-y-6">
          <div class="h-10 bg-slate-800/50 rounded-xl w-3/4"></div>
          <div class="h-6 bg-slate-800/50 rounded-lg w-1/4"></div>
          <div class="h-40 bg-slate-800/50 rounded-xl w-full"></div>
        </div>
      </div>
    </div>

    <!-- Error/404 State -->
    <div
      v-else-if="error || !bike"
      class="text-center py-24 bg-slate-800/20 border border-slate-800 rounded-3xl"
    >
      <div
        class="inline-block p-4 rounded-full bg-red-900/10 text-red-500/80 mb-4"
      >
        <svg
          class="w-12 h-12"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
      </div>
      <h2 class="text-2xl font-bold text-white mb-2">Bike not found</h2>
      <p class="text-slate-500 mb-8 max-w-sm mx-auto">
        The bicycle you are looking for might have been sold or removed from the
        registry.
      </p>
      <NuxtLink
        to="/marketplace"
        class="px-8 py-3 bg-slate-800 hover:bg-slate-700 text-white font-bold rounded-xl transition-all border border-slate-700"
      >
        Return to Marketplace
      </NuxtLink>
    </div>

    <!-- Content -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16">
      <!-- Left: Image Gallery (Compact on desktop) -->
      <div class="lg:col-span-5 space-y-6">
        <div
          class="aspect-square lg:aspect-[4/5] overflow-hidden bg-slate-950 rounded-2xl border border-slate-800 shadow-2xl relative group"
        >
          <img
            :src="activeImage || '/placeholder-bike.jpg'"
            :alt="bike.make_model"
            class="w-full h-full object-cover transition-all duration-500"
          />
          <div class="absolute top-4 left-4">
            <span
              class="px-3 py-1 bg-slate-900/90 backdrop-blur-md border border-slate-700 text-blue-400 text-xs font-bold rounded-lg uppercase tracking-widest shadow-xl"
            >
              {{ bike.status.replace("_", " ") }}
            </span>
          </div>
        </div>

        <!-- Thumbnails -->
        <div v-if="bike.images?.length > 1" class="grid grid-cols-5 gap-3">
          <button
            v-for="img in bike.images"
            :key="img.id"
            @click="activeImage = img.url"
            class="aspect-square rounded-lg overflow-hidden border-2 transition-all duration-200 relative"
            :class="
              activeImage === img.url
                ? 'border-blue-500 ring-4 ring-blue-500/10'
                : 'border-slate-800 hover:border-slate-600'
            "
          >
            <img :src="img.url" class="w-full h-full object-cover" />
            <div
              v-if="activeImage !== img.url"
              class="absolute inset-0 bg-slate-900/20"
            ></div>
          </button>
        </div>
      </div>

      <!-- Right: Details (More space on desktop) -->
      <div class="lg:col-span-7 flex flex-col pt-2">
        <div class="mb-8">
          <div
            class="inline-flex items-center gap-2 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-lg text-blue-400 font-bold tracking-widest uppercase text-xs mb-4"
          >
            <span class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></span>
            {{ bike.year }} MODEL
          </div>
          <h1
            class="text-4xl md:text-5xl font-extrabold text-white mb-4 tracking-tight leading-tight"
          >
            {{ bike.make_model }}
          </h1>
          <div
            class="flex flex-wrap items-center gap-y-2 gap-x-6 text-slate-500 text-sm font-medium"
          >
            <span class="flex items-center gap-1.5">
              <svg
                class="w-4 h-4 text-slate-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
              {{ bike.location_city }}
            </span>
            <span class="flex items-center gap-1.5">
              <svg
                class="w-4 h-4 text-slate-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                />
              </svg>
              REGISTRY ID:
              <span class="text-slate-400 font-mono">{{
                bike.id.split("-")[0].toUpperCase()
              }}</span>
            </span>
          </div>
        </div>

        <div
          class="bg-slate-800/30 border border-slate-800/60 rounded-2xl p-8 mb-10 backdrop-blur-sm relative overflow-hidden"
        >
          <div class="absolute top-0 right-0 p-8 opacity-5">
            <svg
              class="w-24 h-24 text-white"
              fill="currentColor"
              viewBox="0 0 24 24"
            >
              <path d="M12 2L1 21h22L12 2zm0 3.45l8.27 14.3H3.73L12 5.45z" />
            </svg>
          </div>
          <div class="relative z-10">
            <div
              class="text-slate-500 text-xs uppercase font-bold tracking-widest mb-2"
            >
              Market Valuation
            </div>
            <div
              class="text-5xl font-black text-white flex items-baseline gap-2"
            >
              <span class="text-3xl text-slate-500 font-light">$</span
              >{{ bike.price?.toLocaleString() }}
              <span
                class="text-sm text-slate-500 font-bold tracking-tighter uppercase"
                >USD</span
              >
            </div>
          </div>
        </div>

        <div class="mb-12">
          <h3
            class="text-sm font-bold text-slate-500 uppercase tracking-widest mb-4 flex items-center gap-2"
          >
            Description
            <span class="flex-grow h-px bg-slate-800"></span>
          </h3>
          <p class="text-slate-400 leading-relaxed whitespace-pre-wrap text-lg">
            {{
              bike.description || "No description provided for this bicycle."
            }}
          </p>
        </div>

        <!-- Call to Action -->
        <div class="mt-auto space-y-4">
          <button
            class="w-full py-4 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl transition-all shadow-xl shadow-blue-900/20 uppercase tracking-widest"
          >
            Contact Owner
          </button>
          <div
            class="bg-slate-800/50 border border-slate-800 rounded-xl p-4 flex items-center gap-3"
          >
            <div class="p-2 bg-slate-900 rounded-lg text-blue-500">
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                />
              </svg>
            </div>
            <p
              class="text-slate-500 text-xs font-medium uppercase tracking-tight"
            >
              Ownership verified by
              <span class="text-slate-300">VeloTrace Protocol</span>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute();
const bikesApi = useBikesApi();
const routeId = route.params.id;

// 1. Narrow the type
if (Array.isArray(routeId)) {
  throw createError({ statusCode: 400, message: "Invalid ID" });
}

const id = routeId;

const {
  data: bike,
  pending,
  error,
} = await useAsyncData<Bike>(`bike-${route.params.id}`, () =>
  bikesApi
    .GET("/bikes/{id}", {
      params: { path: { id } },
    })
    .then((res) => {
      if (res.error) throw res.error;
      return res.data;
    }),
);

// Gallery State
const activeImage = ref("");

// Initialize active image when data arrives
watch(
  bike,
  (newBike) => {
    if (newBike?.images?.length && !activeImage.value) {
      const primary = newBike.images.find((img) => img.is_primary);
      activeImage.value = primary?.url || newBike.images[0].url;
    }
  },
  { immediate: true },
);

// SEO
useHead({
  title: bike.value ? `${bike.value.make_model} | VeloTrace` : "Bike Details",
  meta: [
    {
      name: "description",
      content: bike.value?.description || "View bicycle details on VeloTrace.",
    },
  ],
});
</script>
