<template>
  <div>
    <!-- 铃铛按钮 -->
    <button
      @click="openModal"
      class="relative flex h-9 w-9 items-center justify-center rounded-sm text-[#756554] transition-colors duration-500 hover:bg-[#d4cdc5]/25 dark:text-[#c9a88c] dark:hover:bg-[#3a3733]"
      :class="{ 'text-[#4f6b5e] dark:text-[#9ab5a6]': unreadCount > 0 }"
      :aria-label="t('announcements.title')"
    >
      <Icon name="bell" size="md" />
      <!-- 未读红点 -->
      <span
        v-if="unreadCount > 0"
        class="absolute right-1 top-1 flex h-2 w-2"
      >
        <span class="absolute inline-flex h-full w-full rounded-sm bg-[#5a7a6b]/20"></span>
        <span class="relative inline-flex h-2 w-2 rounded-sm bg-[#5a7a6b] dark:bg-[#9ab5a6]"></span>
      </span>
    </button>

    <!-- 公告列表 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="isModalOpen"
          class="fixed inset-0 z-[100] flex items-start justify-center overflow-y-auto bg-[#242321]/40 p-4 pt-[8vh]"
          @click="closeModal"
        >
          <div
            class="w-full max-w-[620px] overflow-hidden rounded-lg border border-[#d4cdc5]/40 bg-white shadow-md dark:border-[#4a4540] dark:bg-[#2f2d2a]"
            @click.stop
          >
            <!-- Header with Gradient -->
            <div class="relative overflow-hidden border-b border-[#d4cdc5]/40 bg-[#a89279]/10 px-6 py-5 dark:border-[#4a4540] dark:bg-[#c9a88c]/10">
              <div class="relative z-10 flex items-start justify-between">
                <div>
                  <div class="flex items-center gap-2">
                    <div class="flex h-8 w-8 items-center justify-center rounded-sm bg-[#5a7a6b]/10 text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]">
                      <Icon name="bell" size="sm" />
                    </div>
                    <h2 class="text-lg font-light tracking-wide text-[#3d3d3d] dark:text-[#ede7e0]">
                      {{ t('announcements.title') }}
                    </h2>
                  </div>
                  <p v-if="unreadCount > 0" class="mt-2 text-sm font-light text-[#756554] dark:text-[#c9a88c]">
                    <span class="font-light text-[#4f6b5e] dark:text-[#9ab5a6]">{{ unreadCount }}</span>
                    {{ t('announcements.unread') }}
                  </p>
                </div>
                <div class="flex items-center gap-2">
                  <button
                    v-if="unreadCount > 0"
                    @click="markAllAsRead"
                    :disabled="loading"
                    class="rounded-sm bg-[#5a7a6b] px-4 py-2 text-xs font-light text-[#f5f0eb] transition-colors duration-500 hover:bg-[#4f6b5e] disabled:opacity-50 dark:bg-[#9ab5a6] dark:text-[#2f2d2a] dark:hover:bg-[#b6cddd]"
                  >
                    {{ t('announcements.markAllRead') }}
                  </button>
                  <button
                    @click="closeModal"
                    class="flex h-9 w-9 items-center justify-center rounded-sm text-[#756554] transition-colors duration-500 hover:bg-[#d4cdc5]/25 hover:text-[#3d3d3d] dark:text-[#c9a88c] dark:hover:bg-[#3a3733] dark:hover:text-[#ede7e0]"
                    :aria-label="t('common.close')"
                  >
                    <Icon name="x" size="sm" />
                  </button>
                </div>
              </div>
              <!-- Decorative gradient -->
              <div class="absolute right-0 top-0 h-full w-48 bg-[#a89279]/10 dark:bg-[#c9a88c]/10"></div>
            </div>

            <!-- Body -->
            <div class="max-h-[65vh] overflow-y-auto">
              <!-- Loading -->
              <div v-if="loading" class="flex items-center justify-center py-16">
                <div class="relative">
                  <div class="h-12 w-12 animate-spin rounded-sm border border-[#d4cdc5]/40 border-t-[#5a7a6b] dark:border-[#4a4540] dark:border-t-[#9ab5a6]"></div>
                  <div class="absolute inset-0 h-12 w-12 rounded-sm border border-[#5a7a6b]/10 dark:border-[#9ab5a6]/10"></div>
                </div>
              </div>

              <!-- Announcements List -->
              <div v-else-if="announcements.length > 0">
                <div
                  v-for="item in announcements"
                  :key="item.id"
                  class="group relative flex items-center gap-4 border-b border-[#d4cdc5]/40 px-6 py-4 transition-colors duration-500 hover:bg-[#ede7e0]/45 dark:border-[#4a4540] dark:hover:bg-[#3a3733]/60"
                  :class="{ 'bg-[#5a7a6b]/5 dark:bg-[#9ab5a6]/5': !item.read_at }"
                  style="min-height: 72px"
                  @click="openDetail(item)"
                >
                  <!-- Status Indicator -->
                  <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center">
                    <div
                      v-if="!item.read_at"
                      class="relative flex h-10 w-10 items-center justify-center rounded-sm bg-[#5a7a6b]/10 text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]"
                    >
                      <!-- Pulse ring -->
                      <span class="absolute inline-flex h-full w-full rounded-sm bg-[#5a7a6b]/10 dark:bg-[#9ab5a6]/10"></span>
                      <!-- Icon -->
                      <svg class="relative z-10 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div
                      v-else
                      class="flex h-10 w-10 items-center justify-center rounded-sm bg-[#d4cdc5]/20 text-[#a89279] dark:bg-[#3a3733] dark:text-[#c9a88c]"
                    >
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>

                  <!-- Content -->
                  <div class="flex min-w-0 flex-1 items-center justify-between gap-4">
                    <div class="min-w-0 flex-1">
                      <h3 class="truncate text-sm font-light text-[#3d3d3d] dark:text-[#ede7e0]">
                        {{ item.title }}
                      </h3>
                      <div class="mt-1 flex items-center gap-2">
                        <time class="text-xs font-light text-[#756554] dark:text-[#c9a88c]">
                          {{ formatRelativeTime(item.created_at) }}
                        </time>
                        <span
                          v-if="!item.read_at"
                          class="inline-flex items-center gap-1 rounded-sm bg-[#5a7a6b]/10 px-1.5 py-0.5 text-xs font-light text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]"
                        >
                          <span class="relative flex h-1.5 w-1.5">
                            <span class="absolute inline-flex h-full w-full rounded-sm bg-[#5a7a6b]/20"></span>
                            <span class="relative inline-flex h-1.5 w-1.5 rounded-sm bg-[#5a7a6b] dark:bg-[#9ab5a6]"></span>
                          </span>
                          {{ t('announcements.unread') }}
                        </span>
                      </div>
                    </div>

                    <!-- Arrow -->
                    <div class="flex-shrink-0">
                      <svg
                        class="h-5 w-5 text-[#a89279] dark:text-[#c9a88c]"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>

                  <!-- Unread indicator bar -->
                  <div
                    v-if="!item.read_at"
                    class="absolute left-0 top-0 h-full w-1 bg-[#5a7a6b] dark:bg-[#9ab5a6]"
                  ></div>
                </div>
              </div>

              <!-- Empty State -->
              <div v-else class="flex flex-col items-center justify-center py-16">
                <div class="relative mb-4">
                  <div class="flex h-20 w-20 items-center justify-center rounded-lg bg-[#d4cdc5]/20 dark:bg-[#3a3733]">
                    <Icon name="inbox" size="xl" class="text-[#a89279] dark:text-[#c9a88c]" />
                  </div>
                  <div class="absolute -right-1 -top-1 flex h-6 w-6 items-center justify-center rounded-sm bg-[#5a7a6b]/10 text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]">
                    <svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </div>
                </div>
                <p class="text-sm font-light text-[#3d3d3d] dark:text-[#ede7e0]">{{ t('announcements.empty') }}</p>
                <p class="mt-1 text-xs font-light text-[#756554] dark:text-[#c9a88c]">{{ t('announcements.emptyDescription') }}</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 公告详情 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="detailModalOpen && selectedAnnouncement"
          class="fixed inset-0 z-[110] flex items-start justify-center overflow-y-auto bg-[#242321]/40 p-4 pt-[6vh]"
          @click="closeDetail"
        >
          <div
            class="w-full max-w-[780px] overflow-hidden rounded-lg border border-[#d4cdc5]/40 bg-white shadow-md dark:border-[#4a4540] dark:bg-[#2f2d2a]"
            @click.stop
          >
            <!-- Header with Decorative Elements -->
            <div class="relative overflow-hidden border-b border-[#d4cdc5]/40 bg-[#a89279]/10 px-8 py-6 dark:border-[#4a4540] dark:bg-[#c9a88c]/10">
              <!-- Decorative background elements -->
              <div class="absolute right-0 top-0 h-full w-64 bg-[#a89279]/10 dark:bg-[#c9a88c]/10"></div>
              <div class="absolute -right-8 -top-8 h-32 w-32 rounded-lg bg-[#7ba0b8]/10 dark:bg-[#b6cddd]/10"></div>
              <div class="absolute -left-4 -bottom-4 h-24 w-24 rounded-lg bg-[#c9a88c]/10 dark:bg-[#c9a88c]/10"></div>

              <div class="relative z-10 flex items-start justify-between gap-4">
                <div class="flex-1 min-w-0">
                  <!-- Icon and Category -->
                  <div class="mb-3 flex items-center gap-2">
                    <div class="flex h-10 w-10 items-center justify-center rounded-sm bg-[#5a7a6b]/10 text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]">
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="rounded-sm bg-[#5a7a6b]/10 px-2.5 py-1 text-xs font-light text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]">
                        {{ t('announcements.title') }}
                      </span>
                      <span
                        v-if="!selectedAnnouncement.read_at"
                        class="inline-flex items-center gap-1.5 rounded-sm bg-[#5a7a6b]/10 px-2.5 py-1 text-xs font-light text-[#4f6b5e] dark:bg-[#9ab5a6]/10 dark:text-[#9ab5a6]"
                      >
                        <span class="relative flex h-2 w-2">
                          <span class="absolute inline-flex h-full w-full rounded-sm bg-[#5a7a6b]/20"></span>
                          <span class="relative inline-flex h-2 w-2 rounded-sm bg-[#5a7a6b] dark:bg-[#9ab5a6]"></span>
                        </span>
                        {{ t('announcements.unread') }}
                      </span>
                    </div>
                  </div>

                  <!-- Title -->
                  <h2 class="mb-3 text-2xl font-light leading-tight tracking-wide text-[#3d3d3d] dark:text-[#ede7e0]">
                    {{ selectedAnnouncement.title }}
                  </h2>

                  <!-- Meta Info -->
                  <div class="flex items-center gap-4 text-sm font-light text-[#756554] dark:text-[#c9a88c]">
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      <span>{{ selectedAnnouncement.read_at ? t('announcements.read') : t('announcements.unread') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Close button -->
                <button
                  @click="closeDetail"
                  class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-sm text-[#756554] transition-colors duration-500 hover:bg-[#d4cdc5]/25 hover:text-[#3d3d3d] dark:text-[#c9a88c] dark:hover:bg-[#3a3733] dark:hover:text-[#ede7e0]"
                  :aria-label="t('common.close')"
                >
                  <Icon name="x" size="md" />
                </button>
              </div>
            </div>

            <!-- Body with Enhanced Markdown -->
            <div class="max-h-[60vh] overflow-y-auto bg-white px-8 py-8 dark:bg-[#2f2d2a]">
              <!-- Content with decorative border -->
              <div class="relative">
                <!-- Decorative left border -->
                <div class="absolute left-0 top-0 bottom-0 w-1 rounded-sm bg-[#5a7a6b]/40 dark:bg-[#9ab5a6]/40"></div>

                <div class="pl-6">
                  <div
                    class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                    v-html="renderMarkdown(selectedAnnouncement.content)"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Footer with Actions -->
            <div class="border-t border-[#d4cdc5]/40 bg-[#ede7e0]/45 px-8 py-5 dark:border-[#4a4540] dark:bg-[#242321]/30">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-xs font-light text-[#756554] dark:text-[#c9a88c]">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{{ selectedAnnouncement.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}</span>
                </div>
                <div class="flex items-center gap-3">
                  <button
                    @click="closeDetail"
                    class="rounded-sm border border-[#d4cdc5]/40 bg-white px-5 py-2.5 text-sm font-light text-[#3d3d3d] transition-colors duration-500 hover:border-[#d4cdc5] hover:bg-[#ede7e0]/60 dark:border-[#4a4540] dark:bg-[#2f2d2a] dark:text-[#ede7e0] dark:hover:bg-[#3a3733]"
                  >
                    {{ t('common.close') }}
                  </button>
                  <button
                    v-if="!selectedAnnouncement.read_at"
                    @click="markAsReadAndClose(selectedAnnouncement.id)"
                    class="rounded-sm bg-[#5a7a6b] px-5 py-2.5 text-sm font-light text-[#f5f0eb] transition-colors duration-500 hover:bg-[#4f6b5e] dark:bg-[#9ab5a6] dark:text-[#2f2d2a] dark:hover:bg-[#b6cddd]"
                  >
                    <span class="flex items-center gap-2">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                      {{ t('announcements.markRead') }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore } from '@/stores/app'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import '@/styles/announcement-markdown.css'

const { t } = useI18n()
const appStore = useAppStore()
const announcementStore = useAnnouncementStore()

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true,
})

// Use store state (storeToRefs for reactivity)
const { announcements, loading } = storeToRefs(announcementStore)
const unreadCount = computed(() => announcementStore.unreadCount)

// Local modal state
const isModalOpen = ref(false)
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)

// Methods
function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
}

function openModal() {
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
}

function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) {
    markAsRead(announcement.id)
  }
}

function closeDetail() {
  detailModalOpen.value = false
  selectedAnnouncement.value = null
}

async function markAsRead(id: number) {
  try {
    const marked = await announcementStore.markAsRead(id)
    if (!marked) appStore.showError(t('common.unknownError'))
    return marked
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
    return false
  }
}

async function markAsReadAndClose(id: number) {
  if (!await markAsRead(id)) return
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}

async function markAllAsRead() {
  try {
    await announcementStore.markAllAsRead()
    appStore.showSuccess(t('announcements.allMarkedAsRead'))
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  }
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (detailModalOpen.value) {
      closeDetail()
    } else if (isModalOpen.value) {
      closeModal()
    }
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  document.body.style.overflow = ''
})

watch(
  [isModalOpen, detailModalOpen, () => announcementStore.currentPopup],
  ([modal, detail, popup]) => {
    document.body.style.overflow = (modal || detail || popup) ? 'hidden' : ''
  }
)
</script>

<style scoped>
/* Modal Animations */
.modal-fade-enter-active {
  transition: opacity 0.3s ease;
}

.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
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
