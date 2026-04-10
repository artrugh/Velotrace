<template>
  <div class="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
    <div
      class="mb-10 flex flex-col md:flex-row md:items-end justify-between gap-6"
    >
      <div>
        <h1 class="text-4xl font-extrabold text-white mb-2">Marketplace</h1>
        <p class="text-slate-400 text-lg">
          Verified pre-owned bikes from trusted owners.
        </p>
      </div>
      <NuxtLink
        to="/bikes/register"
        class="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-900/40 text-center"
      >
        Sell Your Bike
      </NuxtLink>
    </div>

    <!-- Loading State: Skeleton Grid -->
    <div
      v-if="pending"
      class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8"
    >
      <div
        v-for="i in 8"
        :key="i"
        class="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden animate-pulse"
      >
        <div class="aspect-[4/3] bg-slate-800"></div>
        <div class="p-5 space-y-4">
          <div class="h-6 bg-slate-800 rounded w-3/4"></div>
          <div class="h-4 bg-slate-800 rounded w-1/2"></div>
          <div
            class="pt-4 flex justify-between items-center border-t border-slate-800"
          >
            <div class="h-6 bg-slate-800 rounded w-1/4"></div>
            <div class="h-10 bg-slate-800 rounded w-1/3"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div
      v-else-if="error"
      class="text-center py-20 bg-slate-900/20 border border-slate-800 rounded-3xl"
    >
      <div
        class="inline-block p-4 rounded-full bg-red-900/20 text-red-500 mb-4"
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
      <h2 class="text-2xl font-bold text-white mb-2">
        Failed to load marketplace
      </h2>
      <p class="text-slate-400 mb-6">
        There was an issue connecting to the bike registry.
      </p>
      <button
        @click="() => refresh()"
        class="px-6 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg transition-colors"
      >
        Try Again
      </button>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="!bikes || bikes.length === 0"
      class="text-center py-32 bg-slate-900/20 border border-slate-800 rounded-3xl"
    >
      <div
        class="inline-block p-4 rounded-full bg-slate-800/50 text-slate-500 mb-4"
      >
        <svg
          class="w-16 h-16"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
          />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M9 22V12h6v10"
          />
        </svg>
      </div>
      <h2 class="text-2xl font-bold text-white mb-2">No bikes for sale yet</h2>
      <p class="text-slate-400 mb-8">
        Be the first to register a bike in our verified marketplace.
      </p>
      <NuxtLink
        to="/bikes/register"
        class="px-8 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl transition-all"
      >
        Register a Bike
      </NuxtLink>
    </div>

    <!-- Bike Grid -->
    <div
      v-else
      class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8"
    >
      <div
        v-for="bike in bikes"
        :key="bike.id"
        class="group bg-slate-900/40 hover:bg-slate-900/60 border border-slate-800 hover:border-blue-500/30 rounded-2xl overflow-hidden transition-all duration-300 hover:shadow-2xl hover:shadow-blue-900/10"
      >
        <!-- Bike Image -->
        <div class="aspect-[4/3] overflow-hidden bg-slate-800 relative">
          <img
            v-if="bike.images?.length"
            :src="getPrimaryImage(bike.images)"
            :alt="bike.make_model"
            class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
          />
          <div
            v-else
            class="w-full h-full flex items-center justify-center text-slate-600"
          >
            <svg
              class="w-16 h-16"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1"
                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </div>
          <div class="absolute top-3 right-3">
            <span
              class="px-3 py-1 bg-slate-950/80 backdrop-blur-md border border-slate-700 text-slate-300 text-xs font-bold rounded-full uppercase tracking-wider"
            >
              Verified
            </span>
          </div>
        </div>

        <!-- Bike Info -->
        <div class="p-5">
          <div
            class="mb-1 text-xs font-bold text-blue-500 uppercase tracking-widest"
          >
            {{ bike.year || "NEW" }}
          </div>
          <h3 class="text-xl font-bold text-white mb-1 truncate">
            {{ bike.make_model }}
          </h3>
          <p class="text-slate-500 text-sm mb-4 flex items-center gap-1">
            <svg
              class="w-4 h-4"
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
            {{ bike.location_city || "Global" }}
          </p>

          <div
            class="pt-4 flex items-center justify-between border-t border-slate-800"
          >
            <div class="text-2xl font-black text-white">
              ${{ bike.price?.toLocaleString() || "0" }}
            </div>
            <button
              class="px-4 py-2 bg-slate-800 group-hover:bg-blue-600 text-white text-sm font-bold rounded-lg transition-colors"
            >
              View Details
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const bikesApi = useBikesApi();

// SSR Data Fetching via BFF Proxy
const {
  data: bikes,
  pending,
  error,
  refresh,
} = await useAsyncData("marketplace-bikes", () =>
  bikesApi.GET("/bikes").then((res) => {
    if (res.error) throw res.error;
    return res.data?.bikes ?? [];
  }),
);
/**
 * Utility to find the primary image or return the first available
 */
const getPrimaryImage = (images: BikeImage[]) => {
  const primary = images.find((img) => img.is_primary);
  return primary?.url || images[0]?.url;
};

// SEO Metadata
useHead({
  title: "Bicycle Marketplace | VeloTrace",
  meta: [
    {
      name: "description",
      content:
        "Browse verified bicycles for sale in our secure registry-first marketplace.",
    },
  ],
});
</script>
