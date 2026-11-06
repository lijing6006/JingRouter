export default {
  resetLog: {
    title: '重置记录',
    description: '订阅 / 上游账号 / API 密钥的用量重置记录，所有登录用户可见。重置前清掉的额度以美元计。',
    empty: '暂无重置记录',
    loadFailed: '加载重置记录失败',
    filters: {
      all: '全部',
      kind: '重置类型',
      actorEmail: '操作者邮箱'
    },
    kinds: {
      subscription: '订阅',
      account: '上游账号',
      api_key: 'API 密钥'
    },
    columns: {
      time: '时间',
      actor: '操作者',
      kind: '重置类型',
      target: '目标',
      windows: '重置窗口',
      amount: '清掉额度',
      status: '状态'
    },
    status: {
      restored: '已自动恢复',
      notRestored: '未恢复'
    },
    windows: {
      daily: '日',
      weekly: '周',
      monthly: '月',
      quota: '配额',
      rate_limit: '限速'
    },
    amounts: {
      daily: '日用量',
      weekly: '周用量',
      monthly: '月用量',
      quota_used: '已用配额',
      quota_daily_used: '当日已用',
      quota_weekly_used: '本周已用',
      usage_5h: '5 小时用量',
      usage_1d: '1 天用量',
      usage_7d: '7 天用量'
    }
  }
}
