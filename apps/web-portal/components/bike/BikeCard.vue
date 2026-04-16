<script setup lang="ts">
import type { Bike } from "~/composables/useApi";

const props = defineProps<{
  bike: Bike;
}>();

const emit = defineEmits<{
  (e: "click", id: string): void;
}>();

const getStatusClass = (status: string) => {
  const baseClass =
    "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium";
  const variants: Record<string, string> = {
    stolen: "bg-red-100 text-red-800",
    for_sale: "bg-green-100 text-green-800",
    registered: "bg-blue-100 text-blue-800",
    transferred: "bg-gray-100 text-gray-800",
  };
  return `${baseClass} ${variants[status] || "bg-gray-100 text-gray-800"}`;
};

const handleClick = () => {
  emit("click", props.bike.id);
};

const primaryImage = computed(
  () => props.bike.images?.find((img) => img.is_primary)?.url,
);
</script>

<template>
  <div
    class="group relative flex flex-col cursor-pointer bg-white rounded-lg p-2 transition-all hover:shadow-md"
    @click="handleClick"
  >
    <div
      class="aspect-h-1 aspect-w-1 w-full overflow-hidden rounded-md bg-gray-200 lg:aspect-none group-hover:opacity-75 lg:h-80"
    >
      <img
        :src="primaryImage"
        :alt="bike.make_model"
        class="h-full w-full object-cover object-center lg:h-full lg:w-full"
      />
    </div>
    <div class="mt-4 flex justify-between">
      <div>
        <h3 class="text-sm font-semibold text-gray-700">
          {{ bike.make_model }}
        </h3>
        <p class="mt-1 text-sm text-gray-500">{{ bike.location_city }}</p>
      </div>
      <p class="text-sm font-bold text-gray-900">${{ bike.price }}</p>
    </div>
    <div class="mt-2">
      <span :class="getStatusClass(bike.status)">
        {{ bike.status }}
      </span>
    </div>
  </div>
</template>
