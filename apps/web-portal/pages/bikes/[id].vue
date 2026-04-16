<script setup lang="ts">
import { useBikeQueries } from "~/composables/useBikeQueries";

const route = useRoute();
const { fetchBikeById } = useBikeQueries();

const id = Array.isArray(route.params.id)
  ? route.params.id[0]
  : route.params.id;

const {
  data: bike,
  pending,
  error,
  refresh,
} = await useAsyncData(`bike-${id}`, () => fetchBikeById(id));

const activeImage = ref("");

const validImages = computed(
  () => bike.value?.images?.filter((img) => !!img.url) ?? [],
);

watch(
  bike,
  (newBike) => {
    if (validImages.value.length && !activeImage.value) {
      const primary = validImages.value.find((img) => img.is_primary);
      activeImage.value = primary?.url || validImages.value[0].url!;
    }
  },
  { immediate: true },
);

useHead({
  title: computed(() =>
    bike.value ? `${bike.value.make_model} | VeloTrace` : "Bike Details",
  ),
});
</script>

<template>
  <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-10">
    <div class="mb-8">
      <NuxtLink
        to="/marketplace"
        class="text-gray-500 hover:text-gray-900 flex items-center gap-2 transition-colors group text-sm font-medium"
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

    <div v-if="pending" class="animate-pulse">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
        <div class="aspect-square bg-gray-200 rounded-2xl"></div>
        <div class="space-y-6">
          <div class="h-10 bg-gray-200 rounded-xl w-3/4"></div>
          <div class="h-40 bg-gray-200 rounded-xl w-full"></div>
        </div>
      </div>
    </div>

    <div
      v-else-if="error || !bike"
      class="text-center py-24 bg-gray-50 border border-gray-200 rounded-3xl"
    >
      <h2 class="text-2xl font-bold text-gray-900 mb-2">Bike not found</h2>
      <p class="text-gray-500 mb-8 max-w-sm mx-auto">
        The bicycle you are looking for might have been sold or removed from the
        registry.
      </p>
      <div class="flex flex-col sm:flex-row items-center justify-center gap-4">
        <button
          @click="refresh"
          class="px-8 py-3 bg-white text-gray-900 font-bold rounded-xl transition-all border border-gray-200 hover:bg-gray-50"
        >
          Try Again
        </button>
        <NuxtLink
          to="/marketplace"
          class="px-8 py-3 bg-indigo-600 text-white font-bold rounded-xl transition-all"
        >
          Return to Marketplace
        </NuxtLink>
      </div>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-16">
      <div class="lg:col-span-5 space-y-6">
        <div
          class="aspect-square overflow-hidden bg-gray-100 rounded-2xl border border-gray-200 shadow-lg relative"
        >
          <img
            :src="activeImage"
            :alt="bike.make_model"
            class="w-full h-full object-cover"
          />
          <div class="absolute top-4 left-4">
            <span
              class="px-3 py-1 bg-white/90 backdrop-blur-sm border border-gray-200 text-indigo-600 text-xs font-bold rounded-lg uppercase tracking-widest shadow-sm"
            >
              {{ bike.status }}
            </span>
          </div>
        </div>

        <div v-if="validImages.length > 1" class="grid grid-cols-5 gap-3">
          <button
            v-for="img in validImages"
            :key="img.id"
            @click="activeImage = img.url!"
            class="aspect-square rounded-lg overflow-hidden border-2 transition-all"
            :class="
              activeImage === img.url
                ? 'border-indigo-500 shadow-md'
                : 'border-transparent hover:border-gray-300'
            "
          >
            <img :src="img.url!" class="w-full h-full object-cover" />
          </button>
        </div>
      </div>

      <div class="lg:col-span-7 flex flex-col">
        <div class="mb-8">
          <div
            class="inline-flex items-center gap-2 px-3 py-1 bg-indigo-50 border border-indigo-100 rounded-lg text-indigo-600 font-bold tracking-widest uppercase text-xs mb-4"
          >
            {{ bike.year }} MODEL
          </div>
          <h1 class="text-4xl font-extrabold text-gray-900 mb-4 tracking-tight">
            {{ bike.make_model }}
          </h1>
          <div
            class="flex flex-wrap items-center gap-y-2 gap-x-6 text-gray-500 text-sm"
          >
            <span class="flex items-center gap-1.5 font-medium">
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
              {{ bike.location_city }}
            </span>
            <span class="font-mono text-xs"
              >REGISTRY ID: {{ bike.id?.split("-")[0].toUpperCase() }}</span
            >
          </div>
        </div>

        <div class="bg-gray-50 border border-gray-200 rounded-2xl p-8 mb-10">
          <div
            class="text-gray-500 text-xs uppercase font-bold tracking-widest mb-2"
          >
            Market Valuation
          </div>
          <div
            class="text-5xl font-black text-gray-900 flex items-baseline gap-2"
          >
            <span class="text-3xl text-gray-400 font-light">$</span
            >{{ bike.price?.toLocaleString() }}
          </div>
        </div>

        <div class="mb-12">
          <h3
            class="text-sm font-bold text-gray-900 uppercase tracking-widest mb-4"
          >
            Description
          </h3>
          <p class="text-gray-600 leading-relaxed whitespace-pre-wrap">
            {{
              bike.description || "No description provided for this bicycle."
            }}
          </p>
        </div>

        <div class="mt-auto space-y-4">
          <button
            class="w-full py-4 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-900/20 uppercase tracking-widest"
          >
            Contact Owner
          </button>
          <div
            class="flex items-center justify-center gap-2 text-gray-500 text-xs font-medium"
          >
            <svg
              class="w-4 h-4 text-green-500"
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
            OWNERSHIP VERIFIED BY VELOTRACE PROTOCOL
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
