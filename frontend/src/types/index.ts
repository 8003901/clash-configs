export type UpdateSchedule = 'DAY' | 'WEEK'

export interface ClashConfig {
  id: string | null
  url: string | null
  name: string | null
  enabled: boolean
  updateSchedule: UpdateSchedule
  content: string | null
  createdAt: string | null
  updatedAt: string | null
  subscriptionUserinfo: string | null
}

export interface ClashConfigsMerge {
  id: string | null
  name: string | null
  token: string | null
  config: string | null
  createdAt: string | null
  updatedAt: string | null
  configs: ClashConfig[] | null
}

export interface ClashConfigAdd {
  url: string
  name: string
  updateSchedule: UpdateSchedule
  enabled: boolean
}

export interface ClashConfigsMergeAdd {
  name: string
  configIds: string[] | null
}

export interface ClashConfigsMergeUpdate {
  name: string
  config: string
  configs: { id: string }[] | null
}

export interface DataUsage {
  upload: number
  download: number
  total: number
  expire: number
}
