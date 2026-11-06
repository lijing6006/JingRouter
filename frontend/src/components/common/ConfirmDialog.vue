<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="handleCancel">
    <div class="space-y-4">
      <p class="text-sm font-light text-[#756554] dark:text-[#c9a88c]">{{ message }}</p>
      <slot></slot>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button
          @click="handleCancel"
          type="button"
          class="rounded-sm border border-[#d4cdc5]/40 bg-white px-4 py-2 text-sm font-light text-[#3d3d3d] transition-colors duration-500 hover:border-[#d4cdc5] hover:bg-[#ede7e0]/60 focus:outline-none focus:ring-1 focus:ring-[#5a7a6b]/30 focus:ring-offset-2 focus:ring-offset-[#f5f0eb] dark:border-[#4a4540] dark:bg-[#2f2d2a] dark:text-[#ede7e0] dark:hover:bg-[#3a3733] dark:focus:ring-offset-[#242321]"
        >
          {{ cancelText }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :class="[
            'rounded-sm px-4 py-2 text-sm font-light text-[#f5f0eb] transition-colors duration-500 focus:outline-none focus:ring-1 focus:ring-offset-2 focus:ring-offset-[#f5f0eb] dark:text-[#2f2d2a] dark:focus:ring-offset-[#242321]',
            danger
              ? 'bg-[#a89279] hover:bg-[#756554] focus:ring-[#a89279]/30 dark:bg-[#c9a88c] dark:hover:bg-[#d7bb9f]'
              : 'bg-[#5a7a6b] hover:bg-[#4f6b5e] focus:ring-[#5a7a6b]/30 dark:bg-[#9ab5a6] dark:hover:bg-[#b6cddd]'
          ]"
        >
          {{ confirmText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'

const { t } = useI18n()

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  danger: false
})

const confirmText = computed(() => props.confirmText || t('common.confirm'))
const cancelText = computed(() => props.cancelText || t('common.cancel'))

const emit = defineEmits<Emits>()

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
}
</script>
