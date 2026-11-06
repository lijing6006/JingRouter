<template>
  <div v-if="entry.status === 'idle'" class="mt-0.5 text-xs">
    <button
      type="button"
      class="text-[#4f6b5e] underline decoration-dashed underline-offset-2 transition-colors duration-500 hover:text-[#3d3d3d] dark:text-[#9ab5a6] dark:hover:text-[#b6cddd]"
      @click="handleFetch"
    >
      {{ t('usage.ipGeo.fetch') }}
    </button>
  </div>

  <div
    v-else-if="entry.status === 'loading'"
    class="mt-0.5 flex items-center gap-1 text-xs font-extralight text-[#756554] dark:text-[#c9a88c]"
  >
    <svg class="h-3 w-3 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
      />
    </svg>
    {{ t('usage.ipGeo.fetching') }}
  </div>

  <div v-else-if="entry.status === 'success'" class="mt-0.5 flex items-center gap-1 text-xs">
    <button
      type="button"
      class="truncate text-[#756554] underline decoration-dotted underline-offset-2 transition-colors duration-500 hover:text-[#4f6b5e] dark:text-[#c9a88c] dark:hover:text-[#9ab5a6]"
      :title="tooltipText"
      @click="handleOpenDetail"
    >
      {{ entry.label }}
    </button>
    <button
      type="button"
      class="text-[#a89279] transition-colors duration-500 hover:text-[#4f6b5e] dark:text-[#c9a88c] dark:hover:text-[#9ab5a6]"
      :title="t('usage.ipGeo.refreshTitle')"
      @click="handleRefresh"
    >
      <Icon name="refresh" size="xs" />
    </button>
  </div>

  <div v-else-if="entry.status === 'error'" class="mt-0.5 text-xs">
    <button
      type="button"
      class="text-[#8b5e52] underline decoration-dashed underline-offset-2 transition-colors duration-500 hover:text-[#756554] dark:text-[#c99b90] dark:hover:text-[#c9a88c]"
      @click="handleFetch"
    >
      {{ t('usage.ipGeo.failed') }}
    </button>
  </div>

  <div v-else class="mt-0.5 text-xs font-extralight text-[#756554] dark:text-[#c9a88c]">
    {{ t('usage.ipGeo.private') }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { fetchOne, getEntry } from '@/utils/ipGeoLookup'

const props = defineProps<{ ip: string }>()
const { t } = useI18n()

const entry = computed(() => getEntry(props.ip))

const tooltipText = computed(() => {
  const detail = entry.value.detail
  if (!detail) return ''
  const lines = [
    detail.organization ? `${t('usage.ipGeo.detailOrg')}: ${detail.organization}` : '',
    detail.timezone ? `${t('usage.ipGeo.detailTimezone')}: ${detail.timezone}` : '',
    detail.accuracy != null ? `${t('usage.ipGeo.detailAccuracy')}: ${detail.accuracy}km` : '',
    detail.latitude && detail.longitude
      ? `${t('usage.ipGeo.detailCoordinates')}: ${detail.latitude}, ${detail.longitude}`
      : '',
  ].filter(Boolean)
  return lines.join('\n')
})

const handleFetch = () => {
  void fetchOne(props.ip)
}

const handleRefresh = () => {
  void fetchOne(props.ip, true)
}

const handleOpenDetail = () => {
  window.open(
    `https://www.iplocation.net/ip-lookup?query=${encodeURIComponent(props.ip)}`,
    '_blank',
    'noopener,noreferrer'
  )
}
</script>
