<script setup lang="ts">
import { useBikeQueries } from "~/composables/useBikeQueries";
import BikeCard from "~/components/bike/BikeCard.vue";

const { fetchMyBikes } = useBikeQueries();
const router = useRouter();

const {
  data: bikes,
  pending,
  error,
  refresh,
} = await useAsyncData("my-bikes", () => fetchMyBikes());

const navigateToBike = (id: string) => {
  router.push(`/bikes/${id}`);
};

useHead({
  title: "My Collection | VeloTrace",
});
</script>

<template>
  <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-10">
    <div
      class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8"
    >
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-gray-900">
          My Collection
        </h1>
        <p class="mt-2 text-sm text-gray-600">
          Manage your registered bicycles and ownership records.
        </p>
      </div>
      <NuxtLink
        to="/bikes/register"
        class="rounded-md bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 transition-colors text-center"
      >
        Register New Bike
      </NuxtLink>
    </div>

    <div v-if="error" class="rounded-md bg-red-50 p-4 mt-6 text-center">
      <h3 class="text-sm font-medium text-red-800">
        Could not load your collection
      </h3>
      <p class="mt-2 text-sm text-red-700">Please ensure you are logged in.</p>
      <button
        @click="refresh"
        class="mt-4 text-sm font-semibold text-red-600 hover:text-red-500"
      >
        Try Again
      </button>
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
      </div>
    </div>

    <div
      v-else
      class="mt-6 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 lg:grid-cols-4 xl:gap-x-8"
    >
      <BikeCard
        v-for="bike in bikes"
        :key="bike.id"
        :bike="bike"
        @click="navigateToBike"
      />
    </div>

    <div
      v-if="!pending && !error && (!bikes || bikes.length === 0)"
      class="mt-20 text-center"
    >
      <h3 class="mt-4 text-sm font-semibold text-gray-900">
        Your garage is empty
      </h3>
      <p class="mt-1 text-sm text-gray-500 italic">
        Start protecting your property by registering your first bicycle.
      </p>
      <NuxtLink
        to="/bikes/register"
        class="mt-6 inline-block text-sm font-semibold text-blue-600 hover:text-blue-500"
      >
        Register My First Bike &rarr;
      </NuxtLink>
    </div>
  </div>
</template>
