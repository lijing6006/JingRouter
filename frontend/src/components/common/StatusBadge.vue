<template>
  <div :class="['inline-flex items-center gap-1.5 rounded-sm px-1.5 py-0.5', variantClass]">
    <span class="inline-block h-2 w-2 rounded-sm bg-current"></span>
    <span class="text-sm font-light">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
  label: string
}>()

const variantClass = computed(() => {
  switch (props.status) {
    case 'active':
    case 'success':
      return 'bg-[#5a7a6b]/10 text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]'
    case 'disabled':
    case 'inactive':
    case 'warning':
      return 'bg-[#7ba0b8]/10 text-[#4f6879] dark:bg-[#b6cddd]/10 dark:text-[#b6cddd]'
    case 'error':
    case 'danger':
      return 'bg-[#c9a88c]/10 text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]'
    default:
      return 'bg-[#d4cdc5]/20 text-[#756554] dark:bg-[#4a4540]/30 dark:text-[#c9a88c]'
  }
})
</script>
