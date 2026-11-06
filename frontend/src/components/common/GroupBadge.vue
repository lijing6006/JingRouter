<template>
  <span
    :class="[
      'inline-flex items-center gap-1.5 rounded-sm border border-[#d4cdc5]/40 px-2 py-0.5 text-xs font-light transition-colors duration-500 dark:border-[#4a4540]/70',
      badgeClass
    ]"
  >
    <!-- Platform logo -->
    <PlatformIcon v-if="platform" :platform="platform" size="sm" />
    <!-- Group name -->
    <span :class="truncate ? 'truncate' : 'min-w-0 whitespace-normal [overflow-wrap:anywhere]'">{{ name }}</span>
    <!-- Right side label -->
    <span v-if="showLabel" :class="labelClass">
      <template v-if="hasCustomRate">
        <!-- 原倍率删除线 + 专属倍率高亮 -->
        <span class="line-through opacity-50 mr-0.5">{{ rateMultiplier }}x</span>
        <span class="font-light">{{ userRateMultiplier }}x</span>
      </template>
      <template v-else>
        {{ labelText }}
      </template>
    </span>
    <span v-if="hasPeakRate" :class="peakRateClass" :title="peakRateTitle">
      {{ peakRateText }}
    </span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionType, GroupPlatform } from '@/types'
import { useAppStore } from '@/stores/app'
import { formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import PlatformIcon from './PlatformIcon.vue'

interface Props {
  name: string
  platform?: GroupPlatform
  subscriptionType?: SubscriptionType
  rateMultiplier?: number
  userRateMultiplier?: number | null // 用户专属倍率
  peakRateEnabled?: boolean
  peakStart?: string
  peakEnd?: string
  peakRateMultiplier?: number
  showRate?: boolean
  /** Allow full group names to wrap in layouts that have room for multiple lines. */
  truncate?: boolean
  daysRemaining?: number | null // 剩余天数（订阅类型时使用）
  /**
   * 订阅分组默认在右侧 label 展示"订阅"或剩余天数；
   * 开启后订阅分组也改为显示倍率（保留订阅主题色 label，配合可用渠道这类
   * 只关心费率、不关心有效期的场景）。
   */
  alwaysShowRate?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  subscriptionType: 'standard',
  showRate: true,
  truncate: true,
  daysRemaining: null,
  userRateMultiplier: null,
  peakRateEnabled: false,
  alwaysShowRate: false
})

const { t } = useI18n()

const isSubscription = computed(() => props.subscriptionType === 'subscription')

// 是否有专属倍率（且与默认倍率不同）
const hasCustomRate = computed(() => {
  return (
    props.userRateMultiplier !== null &&
    props.userRateMultiplier !== undefined &&
    props.rateMultiplier !== undefined &&
    props.userRateMultiplier !== props.rateMultiplier
  )
})

const appStore = useAppStore()

const hasPeakRate = computed(() => {
  return Boolean(props.showRate && props.peakRateEnabled && props.peakStart && props.peakEnd)
})

const peakRateText = computed(() => {
  return formatPeakRateWindow(
    {
      peak_rate_enabled: props.peakRateEnabled,
      peak_start: props.peakStart,
      peak_end: props.peakEnd,
      peak_rate_multiplier: props.peakRateMultiplier
    },
    serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset)
  )
})

const peakRateTitle = computed(() => {
  return t('common.peakRateTooltip', { window: peakRateText.value })
})

// 是否显示右侧标签
const showLabel = computed(() => {
  if (!props.showRate) return false
  // 订阅类型：显示天数或"订阅"
  if (isSubscription.value) return true
  // 标准类型：显示倍率（包括专属倍率）
  return props.rateMultiplier !== undefined || hasCustomRate.value
})

// Label text
const labelText = computed(() => {
  const rateLabel = props.rateMultiplier !== undefined ? `${props.rateMultiplier}x` : ''
  if (isSubscription.value && !props.alwaysShowRate) {
    // 如果有剩余天数，显示天数
    if (props.daysRemaining !== null && props.daysRemaining !== undefined) {
      if (props.daysRemaining <= 0) {
        return t('admin.users.expired')
      }
      return t('admin.users.daysRemaining', { days: props.daysRemaining })
    }
    // 否则显示"订阅"
    return t('groups.subscription')
  }
  return rateLabel
})

// Label style based on type and days remaining
const labelClass = computed(() => {
  const base = 'px-1.5 py-0.5 rounded-sm border border-[#d4cdc5]/40 text-[10px] font-light dark:border-[#4a4540]/70'

  if (!isSubscription.value) {
    // Standard: subtle background (不再为专属倍率使用不同的背景色)
    return `${base} bg-[#d4cdc5]/30 text-[#3d3d3d] dark:bg-[#4a4540]/60 dark:text-[#ede7e0]`
  }

  // 订阅类型：根据剩余天数显示不同颜色
  if (props.daysRemaining !== null && props.daysRemaining !== undefined) {
    if (props.daysRemaining <= 0 || props.daysRemaining <= 3) {
      // 已过期或紧急（<=3天）：红色
      return `${base} bg-[#c9a88c]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]`
    }
    if (props.daysRemaining <= 7) {
      // 警告（<=7天）：橙色
      return `${base} bg-[#a89279]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#ede7e0]`
    }
  }

  // 正常状态或无天数：根据平台显示主题色
  if (props.platform === 'anthropic') {
    return `${base} bg-[#c9a88c]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]`
  }
  if (props.platform === 'openai') {
    return `${base} bg-[#5a7a6b]/10 text-[#3d3d3d] dark:bg-[#9ab5a6]/10 dark:text-[#ede7e0]`
  }
  if (props.platform === 'gemini') {
    return `${base} bg-[#7ba0b8]/10 text-[#3d3d3d] dark:bg-[#b6cddd]/10 dark:text-[#ede7e0]`
  }
  if (props.platform === 'antigravity') {
    return `${base} bg-[#a89279]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]`
  }
  if (props.platform === 'grok') {
    return `${base} bg-[#d4cdc5]/30 text-[#3d3d3d] dark:bg-[#4a4540]/60 dark:text-[#ede7e0]`
  }
  if (props.platform === 'kimi') {
    return `${base} bg-[#c9a88c]/10 text-[#3d3d3d] dark:bg-[#c9a88c]/10 dark:text-[#ede7e0]`
  }
  if (props.platform === 'zhipu') {
    return `${base} bg-[#7ba0b8]/10 text-[#756554] dark:bg-[#b6cddd]/10 dark:text-[#c9a88c]`
  }
  if (props.platform === 'deepseek') {
    return `${base} bg-[#5a7a6b]/10 text-[#756554] dark:bg-[#9ab5a6]/10 dark:text-[#c9a88c]`
  }
  if (props.platform === 'minimax') {
    return `${base} bg-[#a89279]/10 text-[#3d3d3d] dark:bg-[#c9a88c]/10 dark:text-[#ede7e0]`
  }
  if (props.platform === 'composite') {
    return `${base} bg-[#d4cdc5]/30 text-[#756554] dark:bg-[#4a4540]/60 dark:text-[#c9a88c]`
  }
  return `${base} bg-[#7ba0b8]/10 text-[#3d3d3d] dark:bg-[#b6cddd]/10 dark:text-[#ede7e0]`
})

const peakRateClass = computed(() => {
  return 'px-1.5 py-0.5 rounded-sm border border-[#d4cdc5]/40 text-[10px] font-light text-[#756554] bg-[#c9a88c]/10 dark:border-[#4a4540]/70 dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]'
})

// Badge color based on platform and subscription type
const badgeClass = computed(() => {
  if (props.platform === 'anthropic') {
    // Claude: orange theme
    return isSubscription.value
      ? 'bg-[#c9a88c]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]'
      : 'bg-[#c9a88c]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]'
  } else if (props.platform === 'openai') {
    // OpenAI: green theme
    return isSubscription.value
      ? 'bg-[#5a7a6b]/10 text-[#3d3d3d] dark:bg-[#9ab5a6]/10 dark:text-[#ede7e0]'
      : 'bg-[#5a7a6b]/10 text-[#3d3d3d] dark:bg-[#9ab5a6]/10 dark:text-[#ede7e0]'
  }
  if (props.platform === 'gemini') {
    return isSubscription.value
      ? 'bg-[#7ba0b8]/10 text-[#3d3d3d] dark:bg-[#b6cddd]/10 dark:text-[#ede7e0]'
      : 'bg-[#7ba0b8]/10 text-[#3d3d3d] dark:bg-[#b6cddd]/10 dark:text-[#ede7e0]'
  }
  if (props.platform === 'antigravity') {
    return isSubscription.value
      ? 'bg-[#a89279]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]'
      : 'bg-[#a89279]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]'
  }
  if (props.platform === 'grok') {
    return isSubscription.value
      ? 'bg-[#d4cdc5]/30 text-[#3d3d3d] dark:bg-[#4a4540]/60 dark:text-[#ede7e0]'
      : 'bg-[#d4cdc5]/30 text-[#3d3d3d] dark:bg-[#4a4540]/60 dark:text-[#ede7e0]'
  }
  if (props.platform === 'kimi') {
    return isSubscription.value
      ? 'bg-[#c9a88c]/10 text-[#3d3d3d] dark:bg-[#c9a88c]/10 dark:text-[#ede7e0]'
      : 'bg-[#c9a88c]/10 text-[#3d3d3d] dark:bg-[#c9a88c]/10 dark:text-[#ede7e0]'
  }
  if (props.platform === 'zhipu') {
    return isSubscription.value
      ? 'bg-[#7ba0b8]/10 text-[#756554] dark:bg-[#b6cddd]/10 dark:text-[#c9a88c]'
      : 'bg-[#7ba0b8]/10 text-[#756554] dark:bg-[#b6cddd]/10 dark:text-[#c9a88c]'
  }
  if (props.platform === 'deepseek') {
    return isSubscription.value
      ? 'bg-[#5a7a6b]/10 text-[#756554] dark:bg-[#9ab5a6]/10 dark:text-[#c9a88c]'
      : 'bg-[#5a7a6b]/10 text-[#756554] dark:bg-[#9ab5a6]/10 dark:text-[#c9a88c]'
  }
  if (props.platform === 'minimax') {
    return isSubscription.value
      ? 'bg-[#a89279]/10 text-[#3d3d3d] dark:bg-[#c9a88c]/10 dark:text-[#ede7e0]'
      : 'bg-[#a89279]/10 text-[#3d3d3d] dark:bg-[#c9a88c]/10 dark:text-[#ede7e0]'
  }
  if (props.platform === 'composite') {
    return isSubscription.value
      ? 'bg-[#d4cdc5]/30 text-[#756554] dark:bg-[#4a4540]/60 dark:text-[#c9a88c]'
      : 'bg-[#d4cdc5]/30 text-[#756554] dark:bg-[#4a4540]/60 dark:text-[#c9a88c]'
  }
  // Fallback: original colors
  return isSubscription.value
    ? 'bg-[#7ba0b8]/10 text-[#3d3d3d] dark:bg-[#b6cddd]/10 dark:text-[#ede7e0]'
    : 'bg-[#5a7a6b]/10 text-[#3d3d3d] dark:bg-[#9ab5a6]/10 dark:text-[#ede7e0]'
})
</script>
