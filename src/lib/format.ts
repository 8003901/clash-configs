import type { DataUsage } from '@/types'

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']

export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), BYTE_UNITS.length - 1)
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${BYTE_UNITS[i]}`
}

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/** 把 subscription-userinfo 响应头解析成结构化数据（upload/download/total 单位字节，expire 为秒级时间戳） */
export function parseUserInfo(info: string | null | undefined): DataUsage {
  const result: DataUsage = { upload: 0, download: 0, total: 0, expire: 0 }
  if (!info) return result
  for (const pair of info.split(';')) {
    const eq = pair.indexOf('=')
    if (eq < 0) continue
    const key = pair.slice(0, eq).trim()
    const value = Number(pair.slice(eq + 1).trim())
    if (!Number.isFinite(value)) continue
    if (key === 'upload') result.upload = value
    else if (key === 'download') result.download = value
    else if (key === 'total') result.total = value
    else if (key === 'expire') result.expire = value
  }
  return result
}

export function formatExpire(tsSec: number): string {
  if (!tsSec) return '-'
  const date = new Date(tsSec * 1000)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}
