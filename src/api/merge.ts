import type {
  ClashConfigsMerge,
  ClashConfigsMergeAdd,
  ClashConfigsMergeUpdate,
} from '@/types'
import { http } from './http'

export async function listMerges(): Promise<ClashConfigsMerge[]> {
  const { data } = await http.get<ClashConfigsMerge[]>('/clash_configs_merge')
  return data
}

export async function getMerge(id: string): Promise<ClashConfigsMerge> {
  const { data } = await http.get<ClashConfigsMerge>(`/clash_configs_merge/${id}`)
  return data
}

export async function createMerge(payload: ClashConfigsMergeAdd): Promise<ClashConfigsMerge> {
  const { data } = await http.post<ClashConfigsMerge>('/clash_configs_merge', payload)
  return data
}

export async function updateMerge(id: string, payload: ClashConfigsMergeUpdate): Promise<ClashConfigsMerge> {
  const { data } = await http.put<ClashConfigsMerge>(`/clash_configs_merge/${id}`, payload)
  return data
}

export async function refreshMergeToken(id: string): Promise<ClashConfigsMerge> {
  const { data } = await http.put<ClashConfigsMerge>(`/clash_configs_merge/${id}/token`)
  return data
}
