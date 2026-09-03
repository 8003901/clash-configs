import { ref } from 'vue'
import { checkSession, logout as apiLogout } from '@/api/auth'

const authenticated = ref<boolean | null>(null)

export function useAuth() {
  async function ensure(): Promise<boolean> {
    if (authenticated.value === null) {
      authenticated.value = await checkSession()
    }
    return authenticated.value
  }

  function set(value: boolean): void {
    authenticated.value = value
  }

  async function logout(): Promise<void> {
    await apiLogout()
    authenticated.value = false
  }

  return { authenticated, ensure, set, logout }
}
