<script setup lang="ts">
import { useBikeQueries } from "~/composables/useBikeQueries";
import BikeCard from "~/components/bike/BikeCard.vue";

const { fetchMarketplace } = useBikeQueries();
const router = useRouter();

const {
  data: bikes,
  pending,
  error,
  refresh,
} = await useAsyncData("marketplace", () => fetchMarketplace());

const navigateToBike = (id: string) => {
  router.push(`/bikes/${id}`);
};

useHead({
  title: "Bicycle Marketplace | VeloTrace",
  meta: [
    {
      name: "description",
      content: "Browse verified bicycles for sale in the VeloTrace registry.",
    },
  ],
});
</script>

<template>
  <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-10">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-gray-900">
          Bicycle Marketplace
        </h1>
        <p class="mt-2 text-sm text-gray-600">
          Browse verified bicycles in your area.
        </p>
      </div>
      <NuxtLink
        to="/bikes/register"
        class="rounded-md bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 transition-colors"
      >
        Register My Bike
      </NuxtLink>
    </div>

    <div v-if="error" class="rounded-md bg-red-50 p-4 mt-6">
      <div class="flex flex-col items-center">
        <div class="flex items-center">
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-800">
              Failed to load marketplace
            </h3>
            <div class="mt-2 text-sm text-red-700">
              <p>
                {{
                  error.message ||
                  "An unexpected error occurred. Please try again later."
                }}
              </p>
            </div>
          </div>
        </div>
        <button
          @click="() => refresh()"
          class="mt-4 px-4 py-2 bg-white text-red-600 font-semibold rounded-lg border border-red-200 hover:bg-red-50 transition-colors"
        >
          Try Again
        </button>
      </div>
    </div>

    <div
      v-else-if="pending"
      class="mt-10 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 lg:grid-cols-4 xl:gap-x-8"
    >
      <div v-for="i in 4" :key="i" class="animate-pulse">
        <div
          class="aspect-h-1 aspect-w-1 w-full overflow-hidden rounded-md bg-gray-200 lg:h-80"
        />
        <div class="mt-4 h-4 w-3/4 rounded bg-gray-200" />
        <div class="mt-2 h-4 w-1/2 rounded bg-gray-200" />
      </div>
    </div>

    <div
      v-else-if="bikes.length > 0"
      class="mt-6 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 lg:grid-cols-4 xl:gap-x-8"
    >
      <BikeCard
        v-for="bike in bikes"
        :key="bike.id"
        :bike="bike"
        @click="navigateToBike"
      />
    </div>

    <div v-else class="mt-20 text-center">
      <div
        class="inline-flex h-16 w-16 items-center justify-center rounded-full bg-gray-100"
      >
        <svg
          class="h-8 w-8 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M15.75 10.5V6a3.75 3.75 0 1 0-7.5 0v4.5m11.356-1.993 1.263 12c.07.665-.45 1.243-1.119 1.243H4.25a1.125 1.125 0 0 1-1.12-1.243l1.264-12A1.125 1.125 0 0 1 5.513 7.5h12.974c.576 0 1.059.435 1.119 1.007ZM8.625 10.5a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm7.5 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z"
          />
        </svg>
      </div>
      <h3 class="mt-4 text-sm font-semibold text-gray-900">
        No bikes for sale
      </h3>
      <p class="mt-1 text-sm text-gray-500 italic">
        The marketplace is currently empty. Be the first to register a bike!
      </p>
    </div>
  </div>
</template>
