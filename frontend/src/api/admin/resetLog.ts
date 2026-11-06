/**
 * Reset log API.
 *
 * The reset log is visible to every logged-in user (not admin-only): it
 * records who reset what usage (subscription / upstream account / API key),
 * when, and how much was wiped. The read endpoint lives under the user
 * auth route group on the backend; the reset operations themselves stay
 * admin-only.
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export type ResetLogKind = 'subscription' | 'account' | 'api_key'

export interface ResetLog {
  id: number
  created_at: string
  actor_user_id?: number
  actor_email: string
  kind: ResetLogKind
  target_id: number
  target_label: string
  windows: Record<string, boolean>
  amount_before: Record<string, number>
  restored: boolean
  restored_at?: string
  restored_by?: string
}

export interface ResetLogQuery {
  page?: number
  page_size?: number
  kind?: string
  actor_user_id?: number
  actor_email?: string
}

export type ResetLogListResponse = PaginatedResponse<ResetLog>

/**
 * List reset logs (paginated, newest first).
 */
export async function list(params: ResetLogQuery): Promise<ResetLogListResponse> {
  const { data } = await apiClient.get('/admin/reset-logs', { params })
  return data
}

export const resetLogAPI = {
  list
}

export default resetLogAPI
