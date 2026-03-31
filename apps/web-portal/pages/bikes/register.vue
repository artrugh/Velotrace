<template>
  <div class="max-w-4xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
    <div
      class="bg-slate-900/50 backdrop-blur-sm border border-slate-800 rounded-2xl p-8 shadow-xl"
    >
      <div class="mb-8 border-b border-slate-800 pb-6">
        <h1 class="text-3xl font-bold text-white mb-2">Register Your Bike</h1>
        <p class="text-slate-400">
          Establish legal ownership in our secure registry.
        </p>
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-8">
        <!-- Error Alert -->
        <div
          v-if="error"
          class="p-4 bg-red-900/30 border border-red-500/50 rounded-lg text-red-400 text-sm"
        >
          {{ error }}
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Basic Info Section -->
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-1"
                >Make & Model</label
              >
              <input
                v-model="form.make_model"
                type="text"
                placeholder="e.g. Specialized Stumpjumper 2024"
                required
                class="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-1"
                >Serial Number</label
              >
              <input
                v-model="form.serial_number"
                type="text"
                placeholder="Unique frame ID"
                required
                class="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all outline-none"
              />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-slate-300 mb-1"
                  >Year</label
                >
                <input
                  v-model.number="form.year"
                  type="number"
                  placeholder="YYYY"
                  class="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-slate-300 mb-1"
                  >Price (USD)</label
                >
                <input
                  v-model.number="form.price"
                  type="number"
                  placeholder="0.00"
                  class="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none"
                />
              </div>
            </div>
          </div>

          <!-- Description & Location -->
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-1"
                >Location City</label
              >
              <input
                v-model="form.location_city"
                type="text"
                placeholder="Where is the bike located?"
                class="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-1"
                >Description</label
              >
              <textarea
                v-model="form.description"
                rows="4"
                placeholder="Distinguishing features, components, condition..."
                class="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white outline-none"
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Image Upload Section -->
        <div class="border-t border-slate-800 pt-6">
          <label
            class="block text-sm font-medium text-slate-300 mb-3 text-center"
            >Bike Photos (Multiple)</label
          >
          <div
            class="relative border-2 border-dashed border-slate-700 hover:border-blue-500/50 rounded-xl p-8 transition-colors bg-slate-800/20 group cursor-pointer"
            @click="fileInput?.click()"
          >
            <input
              ref="fileInput"
              type="file"
              multiple
              accept="image/*"
              class="hidden"
              @change="handleFileChange"
            />
            <div class="text-center">
              <div
                class="mb-3 text-slate-500 group-hover:text-blue-400 transition-colors"
              >
                <svg
                  class="w-12 h-12 mx-auto"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                  />
                </svg>
              </div>
              <p class="text-slate-400 text-sm">
                Click to upload or drag and drop
              </p>
              <p class="text-slate-600 text-xs mt-1">PNG, JPG up to 10MB</p>
            </div>
          </div>

          <!-- Selected Images Preview -->
          <div
            v-if="selectedImages.length > 0"
            class="mt-4 grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-3"
          >
            <div
              v-for="(img, idx) in selectedImages"
              :key="idx"
              class="relative group aspect-square rounded-lg overflow-hidden border border-slate-700 bg-slate-800"
            >
              <img :src="img.preview" class="w-full h-full object-cover" />
              <button
                type="button"
                @click.stop="removeImage(idx)"
                class="absolute top-1 right-1 p-1 bg-red-600/80 hover:bg-red-600 rounded-full text-white opacity-0 group-hover:opacity-100 transition-opacity"
              >
                <svg
                  class="w-3 h-3"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Progress Overlay -->
        <div
          v-if="isLoading"
          class="fixed inset-0 bg-slate-950/80 backdrop-blur-md z-50 flex items-center justify-center p-6"
        >
          <div class="max-w-md w-full text-center">
            <div class="mb-6">
              <div
                class="inline-block p-4 rounded-full bg-blue-600/20 text-blue-500 mb-4 animate-pulse"
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
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                  />
                </svg>
              </div>
              <h2 class="text-2xl font-bold text-white mb-2">
                Registering your property...
              </h2>
              <p class="text-slate-400">Please do not close this window.</p>
            </div>

            <div
              class="w-full bg-slate-800 rounded-full h-4 overflow-hidden border border-slate-700 mb-3"
            >
              <div
                class="bg-blue-600 h-full transition-all duration-500 ease-out shadow-[0_0_15px_rgba(37,99,235,0.5)]"
                :style="{ width: `${progress}%` }"
              ></div>
            </div>
            <div class="flex justify-between text-xs font-mono">
              <span class="text-slate-500">UPLOADING...</span>
              <span class="text-blue-400 font-bold"
                >{{ Math.round(progress) }}%</span
              >
            </div>
          </div>
        </div>

        <!-- Form Actions -->
        <div
          class="flex items-center justify-end gap-4 pt-4 border-t border-slate-800"
        >
          <NuxtLink
            to="/"
            class="text-slate-400 hover:text-white transition-colors"
            >Cancel</NuxtLink
          >
          <button
            type="submit"
            :disabled="isLoading"
            class="px-8 py-3 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-900/40"
          >
            Complete Registration
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
const { registerBike, isLoading, error, progress } = useBikeRegistration();
const router = useRouter();

const fileInput = ref<HTMLInputElement | null>(null);

const form = reactive({
  make_model: "",
  serial_number: "",
  year: new Date().getFullYear(),
  description: "",
  location_city: "",
  price: 0,
});

const selectedImages = ref<{ file: File; preview: string }[]>([]);

const handleFileChange = (e: Event) => {
  const files = (e.target as HTMLInputElement).files;
  if (!files) return;

  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    selectedImages.value.push({
      file,
      preview: URL.createObjectURL(file),
    });
  }
};

const removeImage = (index: number) => {
  URL.revokeObjectURL(selectedImages.value[index].preview);
  selectedImages.value.splice(index, 1);
};

const handleSubmit = async () => {
  try {
    const files = selectedImages.value.map((img) => img.file);
    await registerBike(form, files);

    // Cleanup previews
    selectedImages.value.forEach((img) => URL.revokeObjectURL(img.preview));

    // Redirect to home or bike list after success
    router.push("/");
  } catch (err) {
    // Error is handled by the composable's reactive state
  }
};

onUnmounted(() => {
  // Ensure we cleanup any leftover object URLs
  selectedImages.value.forEach((img) => URL.revokeObjectURL(img.preview));
});
</script>
