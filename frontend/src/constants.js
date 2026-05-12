export const CONTRACT_TYPES = [
  { value: 'purchase', label: '采购' },
  { value: 'sales', label: '销售' },
  { value: 'service', label: '服务' },
  { value: 'engineering', label: '工程' },
]

export const CONTRACT_STATUS = [
  { value: 'active', label: '履行中' },
  { value: 'completed', label: '已完成' },
  { value: 'terminated', label: '已终止' },
]

export function labelType(v) {
  return CONTRACT_TYPES.find((x) => x.value === v)?.label ?? v
}

export function labelStatus(v) {
  return CONTRACT_STATUS.find((x) => x.value === v)?.label ?? v
}

export function alertLabel(level) {
  if (level === 'red') return '临近/逾期 · 红色'
  if (level === 'orange') return '临近 · 橙色'
  return ''
}
