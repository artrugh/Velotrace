<script setup lang="ts">
defineProps<{
  title?: string;
}>();
</script>

<template>
  <div
    class="overflow-hidden rounded-lg bg-white shadow border border-gray-200"
  >
    <div
      v-if="title || $slots.header"
      class="border-b border-gray-200 px-4 py-5 sm:px-6"
    >
      <slot name="header">
        <h3 class="text-base font-semibold leading-6 text-gray-900">
          {{ title }}
        </h3>
      </slot>
    </div>
    <div class="px-4 py-5 sm:p-6">
      <slot />
    </div>
    <div v-if="$slots.footer" class="bg-gray-50 px-4 py-4 sm:px-6">
      <slot name="footer" />
    </div>
  </div>
</template>
