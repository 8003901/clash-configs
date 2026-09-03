import { http } from './http'

export async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
  await http.put('/users/me/password', { oldPassword, newPassword })
}
