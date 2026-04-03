<template>
  <div class="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
    <div class="mb-10 flex flex-col md:flex-row md:items-end justify-between gap-6">
      <div>
        <div class="flex items-center gap-3 mb-2">
          <div class="p-2 bg-blue-600 rounded-lg text-white shadow-lg shadow-blue-900/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
          <h1 class="text-4xl font-extrabold text-white tracking-tight">My Collection</h1>
        </div>
        <p class="text-slate-400 text-lg">Manage your registered bicycles and ownership records.</p>
      </div>
      <NuxtLink 
        to="/bikes/register" 
        class="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-900/40 text-center flex items-center gap-2"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Register New Bike
      </NuxtLink>
    </div>

    <!-- Loading State -->
    <div v-if="pending" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8">
      <div v-for="i in 4" :key="i" class="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden animate-pulse">
        <div class="aspect-[4/3] bg-slate-800"></div>
        <div class="p-5 space-y-4">
          <div class="h-6 bg-slate-800 rounded w-3/4"></div>
          <div class="h-4 bg-slate-800 rounded w-1/2"></div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="text-center py-20 bg-slate-900/20 border border-slate-800 rounded-3xl">
      <div class="inline-block p-4 rounded-full bg-red-900/20 text-red-500 mb-4">
        <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h2 class="text-2xl font-bold text-white mb-2">Could not load your bikes</h2>
      <p class="text-slate-400 mb-6">Please ensure you are logged in to view your collection.</p>
      <button @click="() => refresh()" class="px-6 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-lg transition-colors">
        Try Again
      </button>
    </div>

    <!-- Empty State -->
    <div v-else-if="!bikes || bikes.length === 0" class="text-center py-32 bg-slate-900/20 border border-slate-800 rounded-3xl">
      <div class="inline-block p-4 rounded-full bg-slate-800/50 text-slate-500 mb-4">
        <svg class="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
      <h2 class="text-2xl font-bold text-white mb-2">Your garage is empty</h2>
      <p class="text-slate-400 mb-8">Start protecting your property by registering your first bicycle.</p>
      <NuxtLink to="/bikes/register" class="px-8 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl transition-all shadow-xl shadow-blue-900/20">
        Register My First Bike
      </NuxtLink>
    </div>

    <!-- Bike Grid -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8">
      <NuxtLink 
        v-for="bike in bikes" 
        :key="bike.id" 
        :to="`/bikes/${bike.id}`"
        class="group bg-slate-900/40 hover:bg-slate-900/60 border border-slate-800 hover:border-blue-500/30 rounded-2xl overflow-hidden transition-all duration-300 hover:shadow-2xl hover:shadow-blue-900/10 block"
      >
        <!-- Bike Image -->
        <div class="aspect-[4/3] overflow-hidden bg-slate-800 relative">
          <img 
            v-if="bike.images?.length"
            :src="getPrimaryImage(bike.images)" 
            :alt="bike.make_model"
            class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
          />
          <div v-else class="w-full h-full flex items-center justify-center text-slate-600 bg-slate-900/50">
            <svg class="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          
          <div class="absolute top-3 left-3">
            <span 
              class="px-3 py-1 backdrop-blur-md border text-[10px] font-black rounded-lg uppercase tracking-widest shadow-xl"
              :class="getStatusClass(bike.status)"
            >
              {{ bike.status.replace('_', ' ') }}
            </span>
          </div>
        </div>

        <!-- Bike Info -->
        <div class="p-5">
          <div class="flex justify-between items-start mb-1">
            <span class="text-xs font-bold text-blue-500 uppercase tracking-widest">{{ bike.year }}</span>
            <span class="text-[10px] font-mono text-slate-600 uppercase tracking-tighter">ID: {{ bike.id.split('-')[0] }}</span>
          </div>
          <h3 class="text-xl font-bold text-white mb-4 group-hover:text-blue-400 transition-colors truncate">
            {{ bike.make_model }}
          </h3>
          
          <div class="flex items-center justify-between pt-4 border-t border-slate-800">
            <div class="flex flex-col">
              <span class="text-[10px] text-slate-500 font-bold uppercase tracking-tight">Status</span>
              <span class="text-white text-sm font-medium">{{ bike.status === 'for_sale' ? 'Listed for Sale' : 'Private Registry' }}</span>
            </div>
            <div class="p-2 bg-slate-800 rounded-lg group-hover:bg-blue-600 text-slate-400 group-hover:text-white transition-all">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </div>
        </div>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
const bikesApi = useBikesApi()

// SSR Data Fetching via BFF Proxy (Points to /api/my/bikes)
const { data: bikes, pending, error, refresh } = await useAsyncData<Bike[]>(
  'my-bikes-collection',
  () => bikesApi.GET('/my/bikes').then(res => {
    if (res.error) throw res.error
    return res.data as Bike[]
  })
)

/**
 * Utility to find the primary image or return the first available
 */
const getPrimaryImage = (images: BikeImage[]) => {
  const primary = images.find(img => img.is_primary)
  return primary?.url || images[0]?.url
}

/**
 * Visual status mapping
 */
const getStatusClass = (status: string) => {
  switch (status) {
    case 'for_sale':
      return 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30'
    case 'stolen':
      return 'bg-red-500/20 text-red-400 border-red-500/30'
    default:
      return 'bg-blue-500/20 text-blue-400 border-blue-500/30'
  }
}

// SEO
useHead({
  title: 'My Bikes | VeloTrace',
  meta: [
    { name: 'description', content: 'Manage your bicycle property and ownership history.' }
  ]
})
</script>
