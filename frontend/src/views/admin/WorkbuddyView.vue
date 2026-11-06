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
            <a
              href="https://wk.jingrouer.top"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary"
              :title="t('admin.workbuddy.addAccountHint')"
            >
              <Icon name="userPlus" size="sm" />
              {{ t('admin.workbuddy.addAccount') }}
            </a>
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

        <!-- Check-in result -->
        <div
          v-if="checkinResult || checkinError"
          class="rounded-3xl border px-5 py-4 text-sm"
          :class="checkinError
            ? 'border-red-200 bg-red-50 text-red-800 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-200'
            : 'border-emerald-200 bg-emerald-50 text-emerald-900 dark:border-emerald-900/60 dark:bg-emerald-950/30 dark:text-emerald-100'"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="font-medium">
                {{ checkinError ? t('admin.workbuddy.checkinFailed') : t('admin.workbuddy.checkinDone') }}
              </p>
              <p v-if="checkinError" class="mt-1">{{ checkinError }}</p>
              <ul v-else-if="checkinResult && checkinResult.length > 0" class="mt-1 space-y-0.5">
                <li v-for="r in checkinResult" :key="r.user || r.day">
                  <span class="font-medium">{{ r.user || '—' }}</span>：{{ checkinOutcomeText(r) }}
                </li>
              </ul>
              <p v-else-if="checkinResult" class="mt-1">{{ t('admin.workbuddy.checkinNoAccounts') }}</p>
              <p class="mt-1.5 text-xs opacity-75">{{ t('admin.workbuddy.checkinCoversAll') }}</p>
            </div>
            <button
              type="button"
              class="flex-shrink-0 rounded-lg px-2 py-0.5 text-base leading-none opacity-60 hover:opacity-100"
              :aria-label="t('common.close')"
              @click="checkinResult = null; checkinError = ''"
            >
              ×
            </button>
          </div>
        </div>

        <!-- Accounts table (same layout as 账号管理) -->
        <div class="overflow-hidden rounded-3xl bg-white shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700">
          <DataTable
            :columns="cols"
            :data="accounts"
            :loading="loading"
            row-key="user"
            default-sort-key="user"
            default-sort-order="asc"
            :estimate-row-height="140"
            :virtualize-threshold="50"
          >
            <!-- 名称：头像字 + 账号名 + 提供方 · 套餐 -->
            <template #cell-user="{ row }">
              <div class="flex items-center gap-2.5">
                <span class="inline-flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-blue-50 text-base font-bold text-blue-600 dark:bg-blue-900/30 dark:text-blue-300">
                  {{ row.user.charAt(0) }}
                </span>
                <div class="min-w-0">
                  <div class="truncate font-medium text-gray-900 dark:text-white">{{ row.user }}</div>
                  <div class="truncate text-xs text-gray-500 dark:text-gray-400">{{ row.name }} · {{ row.plan }}</div>
                </div>
              </div>
            </template>

            <!-- 平台/类型：WorkBuddy 徽章 + 套餐 -->
            <template #cell-platform_type="{ row }">
              <div class="flex min-w-0 flex-col gap-1">
                <span class="inline-flex w-fit items-center rounded-md bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
                  WorkBuddy
                </span>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ row.plan }}</span>
              </div>
            </template>

            <!-- 容量：Total 积分已用 / 上限 -->
            <template #cell-capacity="{ row }">
              <span v-if="totalCreditsOf(row)" class="text-sm text-gray-700 dark:text-gray-300">
                {{ capacityText(row) }}
              </span>
              <span v-else class="text-sm text-gray-400 dark:text-dark-500">—</span>
            </template>

            <!-- 状态：签到开关开着（或状态未知）为正常，开关明确关了为签到已关 -->
            <template #cell-status="{ row }">
              <span
                v-if="isCheckinOff(row)"
                class="inline-flex items-center rounded-md bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
              >
                {{ t('admin.workbuddy.statusCheckinOff') }}
              </span>
              <span
                v-else
                class="inline-flex items-center rounded-md bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
              >
                {{ t('admin.workbuddy.statusNormal') }}
              </span>
            </template>

            <!-- 调度：该账号的每日签到开关（调 Magpie 的签到开关接口） -->
            <template #cell-schedulable="{ row }">
              <button
                type="button"
                :disabled="!row.checkin_key || togglingKey === row.checkin_key"
                :title="checkinToggleTitle(row)"
                class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:focus:ring-offset-dark-800"
                :class="[(row.checkin_enabled ?? false) ? 'bg-primary-500 hover:bg-primary-600' : 'bg-gray-200 hover:bg-gray-300 dark:bg-dark-600 dark:hover:bg-dark-500']"
                @click="handleToggleCheckin(row)"
              >
                <span class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out" :class="[(row.checkin_enabled ?? false) ? 'translate-x-4' : 'translate-x-0']" />
              </button>
            </template>

            <!-- 分组：Magpie-WorkBuddy 上游所绑的分组（每行相同） -->
            <template #cell-groups>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="g in workbuddyGroups"
                  :key="g"
                  class="inline-flex items-center whitespace-nowrap rounded-md bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300"
                >
                  {{ g }}
                </span>
              </div>
            </template>

            <!-- 用量窗口：各积分窗口已用/上限 + 剩余 + 细进度条，末行累计 Tokens（估算） -->
            <template #cell-usage="{ row }">
              <div class="flex min-w-[17rem] flex-col gap-1.5 py-1">
                <div v-for="w in visibleWindows(row)" :key="w.name" class="flex flex-col gap-0.5">
                  <div class="flex items-baseline gap-2 text-xs">
                    <span
                      class="w-24 shrink-0 font-medium text-gray-900 dark:text-white"
                    >
                      {{ shortLabel(w) }}
                    </span>
                    <span class="text-gray-700 dark:text-gray-300">{{ formatNumber(w.amount ?? 0) }} / {{ formatNumber(w.limit) }}</span>
                    <span class="text-gray-500 dark:text-gray-400">{{ t('admin.workbuddy.remaining') }} {{ formatNumber(absoluteRemaining(w)) }}</span>
                  </div>
                  <div class="h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                    <div
                      class="h-full rounded-full"
                      :class="isTotalCredits(w) ? 'bg-emerald-500' : 'bg-indigo-500'"
                      :style="{ width: windowPct(w) + '%' }"
                    />
                  </div>
                </div>
                <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.workbuddy.totalTokensEst') }}: {{ formatTokens(totalTokensOf(row)) }}
                </div>
              </div>
            </template>

            <!-- 操作：签到（一次签全部已开启的账号） -->
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                  :disabled="checkingIn"
                  :title="t('admin.workbuddy.checkinHint')"
                  @click="handleCheckin"
                >
                  <Icon :name="checkingIn ? 'refresh' : 'check'" size="sm" :class="checkingIn ? 'animate-spin' : ''" />
                  <span class="text-xs">{{ checkingIn ? t('admin.workbuddy.checkingIn') : t('admin.workbuddy.checkin') }}</span>
                </button>
                <span v-if="row.read_at" class="hidden text-xs text-gray-400 dark:text-gray-500 xl:inline">
                  {{ formatDateTime(row.read_at) }}
                </span>
              </div>
            </template>
          </DataTable>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { WorkbuddyCheckinResult, WorkbuddyQuotaAccount, WorkbuddyQuotaWindow } from '@/api/admin/workbuddy'
import { extractApiErrorMessage } from '@/utils/apiError'
import { useAppStore } from '@/stores/app'
import AppLayout from '@/components/layout/AppLayout.vue'
import StatCard from '@/components/common/StatCard.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const accounts = ref<WorkbuddyQuotaAccount[]>([])
const fetchedAt = ref(0)
const loading = ref(false)
const error = ref('')
const checkingIn = ref(false)
const checkinResult = ref<WorkbuddyCheckinResult[] | null>(null)
const checkinError = ref('')
const togglingKey = ref<string | null>(null)

/**
 * Magpie-WorkBuddy 上游（账号 id 3）所绑的分组，每行相同：
 * 国模（14）、Workbuddy国模分组（17）、ChatGPT订阅（13）。
 * 分组绑定不常变，若后台改了绑定，这里须同步改。
 */
const workbuddyGroups = ['国模', 'Workbuddy国模分组', 'ChatGPT订阅']

/** 表格列：与 账号管理 保持一致的列名与顺序（无选择框与 ID 列）。 */
const cols = computed<Column[]>(() => [
  { key: 'user', label: t('admin.accounts.columns.name'), sortable: true },
  { key: 'platform_type', label: t('admin.accounts.columns.platformType') },
  { key: 'capacity', label: t('admin.accounts.columns.capacity') },
  { key: 'status', label: t('admin.accounts.columns.status') },
  { key: 'schedulable', label: t('admin.accounts.columns.schedulable') },
  { key: 'groups', label: t('admin.accounts.columns.groups') },
  { key: 'usage', label: t('admin.accounts.columns.usageWindows') },
  { key: 'actions', label: t('admin.accounts.columns.actions') },
])

function isTotalCredits(row: WorkbuddyQuotaWindow): boolean {
  return row.name === 'Total credits'
}

function totalCreditsOf(account: WorkbuddyQuotaAccount): WorkbuddyQuotaWindow | undefined {
  return account.windows.find((w) => w.name === 'Total credits')
}

/** 容量列：Total 积分已用 / 上限。 */
function capacityText(account: WorkbuddyQuotaAccount): string {
  const total = totalCreditsOf(account)
  if (!total) return '—'
  return `${formatNumber(total.amount ?? 0)} / ${formatNumber(total.limit)}`
}

/**
 * 该账号累计 Tokens（估算）= 名字含 "tokens" 的各模型估算窗口 amount 之和。
 * amount 是各模型 tokens 估算的已用量（单位即 token 数）；没有这类窗口时为 0。
 * 这只是估算窗口的加总，不是 Magpie 的官方口径，展示时标注“估算”。
 */
function totalTokensOf(account: WorkbuddyQuotaAccount): number {
  return account.windows.reduce((sum, w) => {
    if (!/tokens/i.test(w.name)) return sum
    return sum + (w.amount ?? 0)
  }, 0)
}

/** Token 数格式化：取整 + 千分位。 */
function formatTokens(value: number): string {
  if (!Number.isFinite(value)) return '—'
  return Math.round(value).toLocaleString()
}

/**
 * 只展示积分窗口：unit 为 credits 的窗口，以及 Total credits。
 * 模型 tokens 估算窗口（如 "deepseek-v4.1-flash tokens (est.)"，无 unit）
 * 不展示——那是估算值，会干扰总额口径。
 */
function visibleWindows(account: WorkbuddyQuotaAccount): WorkbuddyQuotaWindow[] {
  return account.windows.filter((w) => {
    if (/tokens/i.test(w.name)) return false
    return w.unit === 'credits' || isTotalCredits(w)
  })
}

/**
 * 窗口剩余绝对积分 = 上限 - 已用量（amount 缺失时后端 omitempty 不返回，按 0 计），
 * 保留 2 位小数。注意：Magpie 返回的 remaining / used 是 0-100 的百分比口径，
 * 不是积分，绝不能直接相加。
 */
function absoluteRemaining(row: WorkbuddyQuotaWindow): number {
  const remaining = row.limit - (row.amount ?? 0)
  if (!Number.isFinite(remaining)) return 0
  return Math.round(remaining * 100) / 100
}

/**
 * 总剩余额度 = 各账号 Total credits 窗口 (limit - amount) 之和。
 * Total 窗口缺失的账号不计入。
 */
const totalRemaining = computed(() => {
  return accounts.value.reduce((sum, acc) => {
    const total = totalCreditsOf(acc)
    if (!total) return sum
    return sum + absoluteRemaining(total)
  }, 0)
})

/** 总积分已用 = 各账号 Total credits 的 amount（供应商口径已用量）之和。 */
const totalUsed = computed(() => {
  return accounts.value.reduce((sum, acc) => {
    const total = totalCreditsOf(acc)
    return sum + (total && total.amount ? total.amount : 0)
  }, 0)
})

/** 进度条短标签：长窗口名不用 32px 定宽徽章，改用短名 + 自适应宽度。 */
function shortLabel(row: WorkbuddyQuotaWindow): string {
  if (row.name === 'Total credits') return 'Total'
  if (row.name === 'Extra credits') return 'Extra'
  if (row.name === 'Credits') return 'Credits'
  // 模型 tokens 窗口（默认不展示）：去掉 " tokens (est.)" 后缀，用模型短名。
  return row.name.replace(/\s*tokens.*$/i, '').trim() || row.name
}

/** 窗口用量百分比（0-100），由 amount / limit 算出，与展示数字一致。 */
function windowPct(row: WorkbuddyQuotaWindow): number {
  const limit = row.limit || 0
  if (!Number.isFinite(limit) || limit <= 0) return 0
  return Math.min(100, Math.max(0, ((row.amount ?? 0) / limit) * 100))
}

/** 签到开关是否明确关了：key 已知（后端取到了开关状态）且 enabled 为 false。 */
function isCheckinOff(account: WorkbuddyQuotaAccount): boolean {
  return !!account.checkin_key && account.checkin_enabled === false
}

/** 调度开关的悬停提示。 */
function checkinToggleTitle(account: WorkbuddyQuotaAccount): string {
  if (!account.checkin_key) return t('admin.workbuddy.checkinToggleUnknown')
  return account.checkin_enabled
    ? t('admin.workbuddy.checkinToggleOn')
    : t('admin.workbuddy.checkinToggleOff')
}

/**
 * 切换单个账号的每日签到开关（调 Magpie 的签到开关接口）。
 * 只改该账号是否参与每日签到，不会立即触发一次签到。
 * 乐观更新：失败时恢复原状态并弹错。
 */
async function handleToggleCheckin(account: WorkbuddyQuotaAccount): Promise<void> {
  const key = account.checkin_key
  if (!key || togglingKey.value) return
  const next = !(account.checkin_enabled ?? false)
  const prev = account.checkin_enabled
  account.checkin_enabled = next
  togglingKey.value = key
  try {
    await adminAPI.workbuddy.setCheckinEnabled(key, next)
  } catch (err) {
    account.checkin_enabled = prev
    appStore.showError(extractApiErrorMessage(err, t('admin.workbuddy.checkinToggleFailed')))
  } finally {
    togglingKey.value = null
  }
}

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

/** 单个账号签到结果的展示文案。 */
function checkinOutcomeText(r: WorkbuddyCheckinResult): string {
  switch (r.outcome) {
    case 'claimed':
      return (
        t('admin.workbuddy.checkinClaimed', { credit: formatTokens(r.credit ?? 0) }) +
        (r.streak ? ` · ${t('admin.workbuddy.checkinStreak', { n: r.streak })}` : '')
      )
    case 'done':
      return t('admin.workbuddy.checkinAlreadyDone')
    case 'ineligible':
      return t('admin.workbuddy.checkinIneligible')
    case 'inactive':
      return t('admin.workbuddy.checkinInactive')
    case 'captcha':
      return t('admin.workbuddy.checkinCaptcha')
    case 'own-app':
      return t('admin.workbuddy.checkinOwnApp')
    default:
      return r.msg || t('admin.workbuddy.checkinFailed')
  }
}

/**
 * 触发一次 Magpie WorkBuddy 签到。注意：这是覆盖全部已开启账号的一次签到，
 * 不是只签当前行那个号；当天已签过的返回 done。签完自动刷新一次额度。
 * 防重复点击：签到进行中所有行的按钮都禁用。
 */
async function handleCheckin(): Promise<void> {
  if (checkingIn.value) return
  checkingIn.value = true
  checkinError.value = ''
  checkinResult.value = null
  try {
    checkinResult.value = await adminAPI.workbuddy.checkinWorkbuddyAccounts()
  } catch (err) {
    checkinError.value = extractApiErrorMessage(err, t('admin.workbuddy.checkinFailed'), {
      WORKBUDDY_CHECKIN_NOT_CONFIGURED: t('admin.workbuddy.checkinNotConfigured'),
    })
  } finally {
    checkingIn.value = false
    await loadAccounts()
  }
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
