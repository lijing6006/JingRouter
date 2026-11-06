/**
 * 后台左侧菜单显示/隐藏配置。
 *
 * 单一事实来源：可隐藏的 15 个一级管理菜单（与 AppSidebar.vue 的 adminNavItems
 * 对应），「系统设置」本身不在列表里、永远显示，避免把自己锁在外面。
 *
 * 存储：服务端 settings 表，键 `admin_menu_visibility`，值为 JSON 对象
 * { menuKey: visible }；缺失的 key 视为显示（宽容语义：只有显式 false 才隐藏，
 * 与 AppSidebar 的 featureFlag 过滤保持一致）。
 */

export interface AdminMenuVisibilityItem {
  /** 稳定键，与 settings 里存的 key 一致 */
  key: string
  /** 侧边栏文案的 i18n 键（复用 nav.*） */
  labelKey: string
}

/** 可隐藏的一级管理菜单，按侧边栏从上到下排列。 */
export const ADMIN_MENU_VISIBILITY_ITEMS: AdminMenuVisibilityItem[] = [
  { key: 'dashboard', labelKey: 'nav.dashboard' },
  { key: 'ops', labelKey: 'nav.ops' },
  { key: 'users', labelKey: 'nav.users' },
  { key: 'groups', labelKey: 'nav.groups' },
  { key: 'channels', labelKey: 'nav.channelManagement' },
  { key: 'subscriptions', labelKey: 'nav.subscriptions' },
  { key: 'accounts', labelKey: 'nav.accounts' },
  { key: 'workbuddy', labelKey: 'nav.workbuddy' },
  { key: 'announcements', labelKey: 'nav.announcements' },
  { key: 'proxies', labelKey: 'nav.proxies' },
  { key: 'redeem', labelKey: 'nav.redeemCodes' },
  { key: 'promo-codes', labelKey: 'nav.promoCodes' },
  { key: 'orders', labelKey: 'nav.orderManagement' },
  { key: 'usage', labelKey: 'nav.usage' },
  { key: 'audit-logs', labelKey: 'nav.auditLogs' },
]

/** 允许出现在配置里的 key 集合（「settings」不在其中，服务端同样会过滤）。 */
export const ADMIN_MENU_VISIBILITY_KEYS = new Set(
  ADMIN_MENU_VISIBILITY_ITEMS.map((item) => item.key),
)

/**
 * 归一化：只保留已知 key、只保留布尔值；未知 key 与 `settings` 一律丢弃。
 * 输入为 null/undefined/非对象时返回空对象（全部显示）。
 */
export function normalizeAdminMenuVisibility(
  input: unknown,
): Record<string, boolean> {
  const out: Record<string, boolean> = {}
  if (input === null || typeof input !== 'object' || Array.isArray(input)) {
    return out
  }
  for (const [key, value] of Object.entries(input as Record<string, unknown>)) {
    if (typeof value === 'boolean' && ADMIN_MENU_VISIBILITY_KEYS.has(key)) {
      out[key] = value
    }
  }
  return out
}

/** 提交前清洗：与归一化相同语义。 */
export function sanitizeAdminMenuVisibility(
  input: unknown,
): Record<string, boolean> {
  return normalizeAdminMenuVisibility(input)
}

/** 某个菜单是否显示：配置缺失或为 true 都显示，只有显式 false 才隐藏。 */
export function isAdminMenuVisible(
  visibility: Record<string, boolean> | null | undefined,
  key: string,
): boolean {
  return visibility?.[key] !== false
}
