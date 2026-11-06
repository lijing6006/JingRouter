export default {
  resetLog: {
    title: 'Reset Logs',
    description: 'Usage reset records for subscriptions, upstream accounts and API keys. Visible to every logged-in user. Wiped amounts are in USD.',
    empty: 'No reset records yet',
    loadFailed: 'Failed to load reset logs',
    filters: {
      all: 'All',
      kind: 'Reset Type',
      actorEmail: 'Actor Email'
    },
    kinds: {
      subscription: 'Subscription',
      account: 'Upstream Account',
      api_key: 'API Key'
    },
    columns: {
      time: 'Time',
      actor: 'Actor',
      kind: 'Reset Type',
      target: 'Target',
      windows: 'Reset Windows',
      amount: 'Wiped Amount',
      status: 'Status'
    },
    status: {
      restored: 'Auto-restored',
      notRestored: 'Not restored'
    },
    windows: {
      daily: 'Daily',
      weekly: 'Weekly',
      monthly: 'Monthly',
      quota: 'Quota',
      rate_limit: 'Rate limit'
    },
    amounts: {
      daily: 'Daily usage',
      weekly: 'Weekly usage',
      monthly: 'Monthly usage',
      quota_used: 'Used quota',
      quota_daily_used: 'Daily used',
      quota_weekly_used: 'Weekly used',
      usage_5h: '5h usage',
      usage_1d: '1d usage',
      usage_7d: '7d usage'
    }
  }
}
