<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#f5f0eb] px-6 py-20 text-[#3d3d3d] dark:bg-[#242321] dark:text-[#ede7e0] md:py-28">
    <!-- Background -->
    <div class="absolute inset-0 bg-[#f5f0eb] dark:bg-[#242321]"></div>

    <!-- Decorative Elements -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden"></div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-10 text-center">
        <!-- Custom Logo or Default Logo -->
        <template v-if="settingsLoaded">
          <div
            class="mb-5 inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-sm border border-[#d4cdc5]/40 bg-white dark:border-[#4a4540] dark:bg-[#2f2d2a]"
          >
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="mb-2 text-3xl font-light tracking-wide text-[#3d3d3d] dark:text-[#ede7e0]">
            {{ siteName }}
          </h1>
          <p class="text-sm font-light text-[#756554] dark:text-[#c9a88c]">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="rounded-lg border border-[#d4cdc5]/40 bg-white p-8 shadow-sm transition-all duration-500 dark:border-[#4a4540] dark:bg-[#2f2d2a] md:p-10">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-8 text-center text-sm font-light text-[#756554] dark:text-[#c9a88c]">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs font-extralight lowercase tracking-wide text-[#756554] dark:text-[#c9a88c]">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>
