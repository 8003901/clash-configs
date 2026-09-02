import { ref } from 'vue'

const STORAGE_KEY = 'theme'

/** 默认深色（与 clash-verge 一致），以 <html> 上的 .dark 类为唯一事实来源 */
const isDark = ref(typeof document !== 'undefined' && document.documentElement.classList.contains('dark'))

function applyTheme() {
  document.documentElement.classList.toggle('dark', isDark.value)
}

export function useTheme() {
  function toggle() {
    isDark.value = !isDark.value
    localStorage.setItem(STORAGE_KEY, isDark.value ? 'dark' : 'light')
    applyTheme()
  }

  return { isDark, toggle }
}
