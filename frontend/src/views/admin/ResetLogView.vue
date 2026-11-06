<template>
  <AppLayout>
    <TablePageLayout>
      <!-- Filters -->
      <template #filters>
        <div class="card p-4 sm:p-6">
          <div class="flex flex-wrap items-end justify-between gap-4">
            <div class="flex flex-1 flex-wrap items-end gap-4">
              <div class="w-full sm:w-auto sm:min-w-[180px]">
                <label class="input-label">{{ t('admin.resetLog.filters.kind') }}</label>
                <Select v-model="filters.kind" :options="kindOptions" @change="search" />
              </div>

              <div class="w-full sm:w-auto sm:min-w-[220px]">
                <label class="input-label">{{ t('admin.resetLog.filters.actorEmail') }}</label>
                <input v-model.trim="filters.actor_email" type="text" class="input" @keyup.enter="search" />
              </div>
            </div>

            <div class="flex w-full flex-wrap items-center justify-end gap-3 sm:w-auto">
              <button type="button" class="btn btn-primary" :disabled="loading" @click="search">
                {{ t('common.search') }}
              </button>
              <button type="button" class="btn btn-secondary" :disabled="loading" @click="resetFilters">
                {{ t('common.reset') }}
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- Table -->
      <template #table>
        <DataTable :columns="columns" :data="logs" :loading="loading" row-key="id">
          <template #cell-created_at="{ value }">
            <span class="whitespace-nowrap text-gray-600 dark:text-gray-300">{{ formatTime(value) }}</span>
          </template>

          <template #cell-actor="{ row }">
            <div class="min-w-0 max-w-[220px]">
              <div class="truncate font-medium text-gray-900 dark:text-white" :title="row.actor_email">
                {{ row.actor_email || '—' }}
              </div>
              <div v-if="row.actor_user_id" class="mt-0.5 text-xs text-gray-400">ID {{ row.actor_user_id }}</div>
            </div>
          </template>

          <template #cell-kind="{ row }">
            <span class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-semibold text-gray-700 dark:bg-dark-700 dark:text-gray-200">
              {{ kindLabel(row.kind) }}
            </span>
          </template>

          <template #cell-target="{ row }">
            <div class="min-w-0 max-w-[240px]">
              <div class="truncate font-medium text-gray-900 dark:text-white" :title="row.target_label">
                {{ row.target_label || '—' }}
              </div>
              <div class="mt-0.5 text-xs text-gray-400">#{{ row.target_id }}</div>
            </div>
          </template>

          <template #cell-windows="{ row }">
            <div class="flex flex-wrap gap-1">
              <span
                v-for="key in windowKeys(row.windows)"
                :key="key"
                class="inline-flex items-center rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
              >
                {{ windowLabel(key) }}
              </span>
              <span v-if="windowKeys(row.windows).length === 0" class="text-gray-400">—</span>
            </div>
          </template>

          <template #cell-amount="{ row }">
            <div class="min-w-0 max-w-[260px] space-y-0.5">
              <div
                v-for="(entry, index) in amountEntries(row.amount_before)"
                :key="index"
                class="whitespace-nowrap text-sm text-gray-700 dark:text-gray-200"
                :title="String(entry.value)"
              >
                {{ entry.label }} ${{ formatAmount(entry.value) }}
              </div>
              <span v-if="amountEntries(row.amount_before).length === 0" class="text-gray-400">—</span>
            </div>
          </template>

          <template #cell-status="{ row }">
            <span v-if="row.restored" :class="restoredBadgeClass" :title="restoredTitle(row)">
              <span class="h-1.5 w-1.5 rounded-full bg-green-500"></span>
              {{ t('admin.resetLog.status.restored') }}
            </span>
            <span v-else :class="notRestoredBadgeClass">
              <span class="h-1.5 w-1.5 rounded-full bg-gray-400"></span>
              {{ t('admin.resetLog.status.notRestored') }}
            </span>
          </template>

          <template #empty>
            <div class="flex flex-col items-center py-8">
              <Icon name="shield" size="xl" class="mb-4 h-12 w-12 text-gray-300 dark:text-dark-600" />
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('admin.resetLog.empty') }}</p>
            </div>
          </template>
        </DataTable>
      </template>

      <!-- Pagination -->
      <template #pagination>
        <Pagination
          v-if="total > 0"
          :total="total"
          :page="page"
          :page-size="pageSize"
          @update:page="onPageChange"
          @update:pageSize="onPageSizeChange"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type ResetLog } from '@/api/admin'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { Column } from '@/components/common/types'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const logs = ref<ResetLog[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filters = reactive({
  kind: '',
  actor_email: ''
})

const columns = computed<Column[]>(() => [
  { key: 'created_at', label: t('admin.resetLog.columns.time') },
  { key: 'actor', label: t('admin.resetLog.columns.actor') },
  { key: 'kind', label: t('admin.resetLog.columns.kind') },
  { key: 'target', label: t('admin.resetLog.columns.target') },
  { key: 'windows', label: t('admin.resetLog.columns.windows') },
  { key: 'amount', label: t('admin.resetLog.columns.amount') },
  { key: 'status', label: t('admin.resetLog.columns.status') }
])

const kindOptions = computed(() => [
  { value: '', label: t('admin.resetLog.filters.all') },
  { value: 'subscription', label: t('admin.resetLog.kinds.subscription') },
  { value: 'account', label: t('admin.resetLog.kinds.account') },
  { value: 'api_key', label: t('admin.resetLog.kinds.api_key') }
])

function kindLabel(kind: string): string {
  const found = kindOptions.value.find((o) => o.value === kind)
  return found && found.value ? found.label : kind
}

function windowKeys(windows: Record<string, boolean> | undefined): string[] {
  if (!windows) return []
  return Object.keys(windows).filter((key) => windows[key])
}

function windowLabel(key: string): string {
  const label = t(`admin.resetLog.windows.${key}`)
  return label === `admin.resetLog.windows.${key}` ? key : label
}

function amountEntries(amountBefore: Record<string, number> | undefined): { label: string; value: number }[] {
  if (!amountBefore) return []
  return Object.keys(amountBefore).map((key) => {
    const label = t(`admin.resetLog.amounts.${key}`)
    return {
      label: label === `admin.resetLog.amounts.${key}` ? key : label,
      value: amountBefore[key]
    }
  })
}

function formatAmount(value: number): string {
  if (typeof value !== 'number' || Number.isNaN(value)) return '0.00'
  return value.toFixed(2)
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

const restoredBadgeClass =
  'inline-flex items-center gap-1.5 rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-semibold text-green-700 dark:bg-green-900/30 dark:text-green-300'
const notRestoredBadgeClass =
  'inline-flex items-center gap-1.5 rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300'

function restoredTitle(row: ResetLog): string {
  if (row.restored_at) return formatTime(row.restored_at)
  return ''
}

function buildQuery() {
  return {
    page: page.value,
    page_size: pageSize.value,
    kind: filters.kind || undefined,
    actor_email: filters.actor_email || undefined
  }
}

async function fetchLogs() {
  loading.value = true
  try {
    const res = await adminAPI.resetLog.list(buildQuery())
    logs.value = res.items
    total.value = res.total
  } catch (err: any) {
    appStore.showError(err?.message || t('admin.resetLog.loadFailed'))
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  fetchLogs()
}

function resetFilters() {
  filters.kind = ''
  filters.actor_email = ''
  search()
}

function onPageChange(p: number) {
  page.value = p
  fetchLogs()
}

function onPageSizeChange(ps: number) {
  pageSize.value = ps
  page.value = 1
  fetchLogs()
}

onMounted(fetchLogs)
</script>
