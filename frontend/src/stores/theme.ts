import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type Theme = 'light' | 'dark'

const STORAGE_KEY = 'workevent:theme'

function readStored(): Theme | null {
  try {
    const v = localStorage.getItem(STORAGE_KEY)
    return v === 'light' || v === 'dark' ? v : null
  } catch {
    return null // modo privado / storage bloqueado
  }
}

function systemPrefersDark(): boolean {
  return typeof matchMedia === 'function' && matchMedia('(prefers-color-scheme: dark)').matches
}

export const useThemeStore = defineStore('theme', () => {
  const theme = ref<Theme>(readStored() ?? (systemPrefersDark() ? 'dark' : 'light'))

  function apply(t: Theme) {
    document.documentElement.classList.toggle('dark', t === 'dark')
  }

  watch(
    theme,
    (t) => {
      apply(t)
      try {
        localStorage.setItem(STORAGE_KEY, t)
      } catch {
        /* preferência apenas não persiste */
      }
    },
    { immediate: true, flush: 'sync' },
  )

  function toggle() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  return { theme, toggle }
})
