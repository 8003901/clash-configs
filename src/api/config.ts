import type { ClashConfig, ClashConfigAdd } from '@/types'
import { http } from './http'

export async function listConfigs(): Promise<ClashConfig[]> {
  const { data } = await http.get<ClashConfig[]>('/clash_configs')
  return data
}

export async function getConfig(id: string, renew = false): Promise<ClashConfig> {
  const { data } = await http.get<ClashConfig>(`/clash_configs/${id}`, { params: { renew } })
  return data
}

export async function createConfig(payload: ClashConfigAdd): Promise<ClashConfig> {
  const { data } = await http.post<ClashConfig>('/clash_configs', payload)
  return data
}

export async function updateConfig(id: string, payload: ClashConfigAdd): Promise<ClashConfig> {
  const { data } = await http.put<ClashConfig>(`/clash_configs/${id}`, payload)
  return data
}

export async function deleteConfig(id: string): Promise<void> {
  await http.delete(`/clash_configs/${id}`)
}
