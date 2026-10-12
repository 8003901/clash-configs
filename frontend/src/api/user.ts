import { http } from './http'

export interface UserSummary {
  id: string
  username: string
}

export interface CurrentUser {
  username: string
  isAdmin: boolean
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
  await http.put('/users/me/password', { oldPassword, newPassword })
}

export async function getCurrentUser(): Promise<CurrentUser> {
  const { data } = await http.get<CurrentUser>('/users/me')
  return data
}

export async function listUsers(): Promise<UserSummary[]> {
  const { data } = await http.get<UserSummary[]>('/users')
  return data
}

export async function createUser(username: string, password: string): Promise<UserSummary> {
  const { data } = await http.post<UserSummary>('/users', { username, password })
  return data
}
