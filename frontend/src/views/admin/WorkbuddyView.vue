<template>
  <AppLayout>
    <div class="w-full min-w-0 space-y-6 pb-8">
      <!-- Page header -->
      <header
        class="page-header mb-0 rounded-3xl bg-white p-5 shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700 sm:p-6"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="min-w-0">
            <h1 class="page-title flex items-center gap-2 text-xl font-black text-gray-900 dark:text-white">
              <span class="inline-flex h-8 w-8 items-center justify-center rounded-xl bg-blue-50 text-blue-500 dark:bg-blue-900/30 dark:text-blue-400">
                <Icon name="creditCard" size="sm" />
              </span>
              {{ t('admin.workbuddy.title') }}
            </h1>
            <p class="page-description mt-1.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.workbuddy.description') }}
            </p>
            <p v-if="fetchedAt" class="mt-1.5 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.workbuddy.updatedAt') }} {{ formatDateTime(fetchedAt * 1000) }}
            </p>
          </div>
          <div class="flex flex-shrink-0 items-center gap-2">
            <button
              type="button"
              class="btn btn-primary"
              :disabled="loading"
              @click="loadAccounts"
            >
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
              {{ loading ? t('admin.workbuddy.refreshing') : t('admin.workbuddy.refresh') }}
            </button>
          </div>
        </div>
      </header>

      <!-- Loading -->
      <div
        v-if="loading && accounts.length === 0"
        class="flex min-h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
      >
        {{ t('common.loading') }}
      </div>

      <!-- Error -->
      <div
        v-else-if="error"
        class="rounded-3xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-800 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-200"
      >
        <p class="font-medium">{{ t('admin.workbuddy.loadFailed') }}</p>
        <p class="mt-1">{{ error }}</p>
      </div>

      <!-- Empty -->
      <div
        v-else-if="accounts.length === 0"
        class="flex min-h-56 flex-col items-center justify-center rounded-3xl border border-dashed border-gray-300 px-6 text-center dark:border-dark-600"
      >
        <Icon name="users" size="xl" class="text-gray-400" />
        <p class="mt-3 font-medium text-gray-800 dark:text-gray-200">
          {{ t('admin.workbuddy.empty') }}
        </p>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('admin.workbuddy.emptyHint') }}
        </p>
      </div>

      <!-- Content -->
      <template v-else>
        <!-- Summary stats -->
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <StatCard
            :title="t('admin.workbuddy.accountCount')"
            :value="accounts.length"
            icon-variant="primary"
          />
          <StatCard
            :title="t('admin.workbuddy.totalRemaining')"
            :value="formatNumber(totalRemaining)"
            icon-variant="success"
          />
          <StatCard
            :title="t('admin.workbuddy.totalUsed')"
            :value="formatNumber(totalUsed)"
            icon-variant="warning"
          />
        </div>

        <!-- One card per account -->
        <section
          v-for="account in accounts"
          :key="account.user"
          class="overflow-hidden rounded-3xl bg-white shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700"
        >
          <div class="flex flex-wrap items-center gap-3 border-b border-gray-100 px-5 py-4 dark:border-dark-700">
            <span class="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-blue-50 text-base font-bold text-blue-600 dark:bg-blue-900/30 dark:text-blue-300">
              {{ account.user.charAt(0) }}
            </span>
            <div class="min-w-0 flex-1">
              <h2 class="truncate text-base font-semibold text-gray-900 dark:text-white">
                {{ account.user }}
              </h2>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ account.name }} · {{ account.plan }}
              </p>
            </div>
            <span
              v-if="totalCreditsOf(account)"
              class="inline-flex items-center rounded-full bg-blue-50 px-3 py-1 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
            >
              {{ t('admin.workbuddy.totalCredits') }}:
              {{ totalCreditsOf(account)?.display || '—' }}
            </span>
            <span class="text-xs text-gray-400 dark:text-gray-500">
              {{ account.read_at ? formatDateTime(account.read_at) : '' }}
            </span>
          </div>

          <DataTable
            :columns="windowColumns"
            :data="account.windows"
            :row-key="(row: WorkbuddyQuotaWindow) => row.name"
          >
            <template #cell-window="{ row }">
              <span
                class="font-medium text-gray-900 dark:text-white"
                :class="isTotalCredits(row) ? 'font-bold' : ''"
              >
                {{ row.name }}
              </span>
              <span
                v-if="isTotalCredits(row)"
                class="ml-2 inline-flex items-center rounded-md bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
              >
                {{ t('admin.workbuddy.totalCredits') }}
              </span>
            </template>

            <template #cell-used_limit="{ row }">
              <span class="text-sm text-gray-900 dark:text-gray-100">{{ row.display || formatNumber(row.used) }}</span>
            </template>

            <template #cell-remaining="{ row }">
              <span class="text-sm text-gray-900 dark:text-gray-100">{{ formatNumber(row.remaining) }}</span>
            </template>

            <template #cell-expiry="{ row }">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ expiryOf(row) }}</span>
            </template>

            <template #empty>
              <EmptyState
                :title="t('admin.workbuddy.empty')"
                :description="t('admin.workbuddy.emptyHint')"
              />
            </template>
          </DataTable>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { WorkbuddyQuotaAccount, WorkbuddyQuotaWindow } from '@/api/admin/workbuddy'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import StatCard from '@/components/common/StatCard.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const accounts = ref<WorkbuddyQuotaAccount[]>([])
const fetchedAt = ref(0)
const loading = ref(false)
const error = ref('')

const windowColumns: Column[] = [
  { key: 'window', label: t('admin.workbuddy.window'), sortable: false },
  { key: 'used_limit', label: t('admin.workbuddy.usedLimit'), sortable: false },
  { key: 'remaining', label: t('admin.workbuddy.remaining'), sortable: false },
  { key: 'expiry', label: t('admin.workbuddy.expiresAt'), sortable: false },
]

function isTotalCredits(row: WorkbuddyQuotaWindow): boolean {
  return row.name === 'Total credits'
}

function totalCreditsOf(account: WorkbuddyQuotaAccount): WorkbuddyQuotaWindow | undefined {
  return account.windows.find((w) => w.name === 'Total credits')
}

/** 总剩余额度 = 各账号 Total credits 的 remaining 之和。 */
const totalRemaining = computed(() => {
  return accounts.value.reduce((sum, acc) => {
    const total = totalCreditsOf(acc)
    return sum + (total ? total.remaining : 0)
  }, 0)
})

/** 总积分已用 = 各账号 Total credits 的 amount（供应商口径已用量）之和。 */
const totalUsed = computed(() => {
  return accounts.value.reduce((sum, acc) => {
    const total = totalCreditsOf(acc)
    return sum + (total && total.amount ? total.amount : 0)
  }, 0)
})

/** 数字格式化：最多 2 位小数，去掉多余的 0。 */
function formatNumber(value: number): string {
  if (!Number.isFinite(value)) return '—'
  const rounded = Math.round(value * 100) / 100
  return rounded.toLocaleString(undefined, { maximumFractionDigits: 2 })
}

/** 时间格式化为本地 YYYY-MM-DD HH:mm。 */
function formatDateTime(value: string | number): string {
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** 到期时间：优先重置时间，其次预计耗尽时间。 */
function expiryOf(row: WorkbuddyQuotaWindow): string {
  const raw = row.resets_at || row.runs_out_at || ''
  if (!raw) return t('admin.workbuddy.noExpiry')
  return formatDateTime(raw)
}

async function loadAccounts(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const result = await adminAPI.workbuddy.getWorkbuddyAccounts()
    accounts.value = result.accounts || []
    fetchedAt.value = result.fetched_at || 0
  } catch (err) {
    error.value = extractApiErrorMessage(err, t('admin.workbuddy.loadFailed'), {
      WORKBUDDY_NOT_CONFIGURED: t('admin.workbuddy.notConfigured'),
    })
    accounts.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadAccounts()
})
</script>
