/**
 * Admin WorkBuddy accounts API endpoints.
 * Quota snapshots for all WorkBuddy accounts, fetched by the backend from the
 * local Magpie gateway (GET /v1/magpie/quotas). The Magpie gateway key never
 * leaves the backend: the frontend only calls this Sub2API endpoint.
 */

import { apiClient } from '../client'

/** 单个额度窗口（Credits / Extra credits / Total credits / 各模型 tokens 估算）。 */
export interface WorkbuddyQuotaWindow {
  name: string
  used: number
  remaining: number
  limit: number
  display?: string
  unit?: string
  amount?: number
  resets_at?: string
  runs_out_at?: string
  runs_out?: number
  holds_why?: string
}

/** 单个 WorkBuddy 账号的额度快照。 */
export interface WorkbuddyQuotaAccount {
  provider: string
  name: string
  plan: string
  user: string
  kind?: string
  read_at?: string
  windows: WorkbuddyQuotaWindow[]
}

/** 全部 WorkBuddy 账号的额度快照（对齐后端 WorkbuddyQuotaResult）。 */
export interface WorkbuddyQuotaResult {
  accounts: WorkbuddyQuotaAccount[]
  fetched_at: number
}

/** 拉取全部 WorkBuddy 账号的额度快照。 */
export async function getWorkbuddyAccounts(): Promise<WorkbuddyQuotaResult> {
  const { data } = await apiClient.get<WorkbuddyQuotaResult>('/admin/workbuddy/accounts')
  return data
}

export default {
  getWorkbuddyAccounts
}
