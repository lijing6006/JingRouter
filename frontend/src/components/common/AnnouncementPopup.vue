<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="displayedAnnouncement"
        class="fixed inset-0 z-[120] flex items-start justify-center overflow-y-auto bg-[#242321]/40 p-4 pt-[8vh]"
      >
        <div
          class="w-full max-w-[680px] overflow-hidden rounded-lg border border-[#d4cdc5]/40 bg-white shadow-md dark:border-[#4a4540] dark:bg-[#2f2d2a]"
          @click.stop
        >
          <!-- Header with warm gradient -->
          <div class="relative overflow-hidden border-b border-[#d4cdc5]/40 bg-[#a89279]/10 px-8 py-6 dark:border-[#4a4540] dark:bg-[#c9a88c]/10">
            <!-- Decorative background -->
            <div class="absolute right-0 top-0 h-full w-64 bg-[#a89279]/10 dark:bg-[#c9a88c]/10"></div>
            <div class="absolute -right-8 -top-8 h-32 w-32 rounded-lg bg-[#c9a88c]/10 dark:bg-[#c9a88c]/10"></div>
            <div class="absolute -left-4 -bottom-4 h-24 w-24 rounded-lg bg-[#7ba0b8]/10 dark:bg-[#b6cddd]/10"></div>

            <div class="relative z-10">
              <!-- Icon and badge -->
              <div class="mb-3 flex items-center gap-2">
                <div class="flex h-10 w-10 items-center justify-center rounded-sm bg-[#5a7a6b]/10 text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                  </svg>
                </div>
                <span class="inline-flex items-center gap-1.5 rounded-sm bg-[#c9a88c]/10 px-2.5 py-1 text-xs font-light text-[#756554] dark:bg-[#c9a88c]/10 dark:text-[#c9a88c]">
                  <span class="relative flex h-2 w-2">
                    <span class="absolute inline-flex h-full w-full rounded-sm bg-[#c9a88c]/20"></span>
                    <span class="relative inline-flex h-2 w-2 rounded-sm bg-[#c9a88c]"></span>
                  </span>
                  {{ t('announcements.unread') }}
                </span>
              </div>

              <!-- Title -->
              <h2 class="mb-2 text-2xl font-light leading-tight tracking-wide text-[#3d3d3d] dark:text-[#ede7e0]">
                {{ displayedAnnouncement.title }}
              </h2>

              <!-- Time -->
              <div class="flex items-center gap-1.5 text-sm font-light text-[#756554] dark:text-[#c9a88c]">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <time>{{ formatRelativeWithDateTime(displayedAnnouncement.created_at) }}</time>
              </div>
            </div>
          </div>

          <!-- Body -->
          <div class="max-h-[50vh] overflow-y-auto bg-white px-8 py-8 dark:bg-[#2f2d2a]">
            <div class="relative">
              <div class="absolute left-0 top-0 bottom-0 w-1 rounded-sm bg-[#c9a88c]/50 dark:bg-[#c9a88c]/50"></div>
              <div class="pl-6">
                <div
                  class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                  v-html="renderedContent"
                ></div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="border-t border-[#d4cdc5]/40 bg-[#ede7e0]/45 px-8 py-5 dark:border-[#4a4540] dark:bg-[#242321]/30">
            <div class="flex items-center justify-end">
              <button
                @click="handleDismiss"
                data-testid="announcement-popup-dismiss"
                class="rounded-sm bg-[#5a7a6b] px-6 py-2.5 text-sm font-light text-[#f5f0eb] transition-colors duration-500 hover:bg-[#4f6b5e] dark:bg-[#9ab5a6] dark:text-[#2f2d2a] dark:hover:bg-[#b6cddd]"
              >
                <span class="flex items-center gap-2">
                  <svg v-if="preview" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  {{ preview ? t('common.close') : t('announcements.markRead') }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import type { Announcement, UserAnnouncement } from '@/types'
import '@/styles/announcement-markdown.css'

type PreviewAnnouncement = Pick<Announcement | UserAnnouncement, 'title' | 'content' | 'created_at'>

const props = withDefaults(defineProps<{
  announcement?: PreviewAnnouncement | null
  preview?: boolean
}>(), {
  announcement: null,
  preview: false,
})

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const displayedAnnouncement = computed(() => (
  props.preview ? props.announcement : announcementStore.currentPopup
))

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = displayedAnnouncement.value?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  if (props.preview) {
    emit('close')
    return
  }
  announcementStore.dismissPopup()
}

// Manage body overflow — only set, never unset (bell component handles restore)
watch(
  displayedAnnouncement,
  (popup) => {
    if (popup) {
      document.body.style.overflow = 'hidden'
    } else if (props.preview) {
      document.body.style.overflow = ''
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  if (props.preview) {
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.popup-fade-enter-active {
  transition: opacity 0.3s ease;
}

.popup-fade-leave-active {
  transition: opacity 0.2s ease;
}

.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #d4cdc5;
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: #4a4540;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #a89279;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #c9a88c;
}
</style>
