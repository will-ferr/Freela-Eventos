import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastKind = 'success' | 'error' | 'info'

export interface Toast {
  id: number
  kind: ToastKind
  title: string
  description?: string
}

const MAX_VISIBLE = 4
const DURATION_MS: Record<ToastKind, number> = { success: 4000, info: 5000, error: 8000 }

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])
  const timers = new Map<number, ReturnType<typeof setTimeout>>()
  let seq = 0

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
    const timer = timers.get(id)
    if (timer) clearTimeout(timer)
    timers.delete(id)
  }

  function push(kind: ToastKind, title: string, description?: string): number {
    const id = ++seq
    toasts.value.push({ id, kind, title, description })
    // Mantém a pilha curta: o mais antigo sai primeiro.
    while (toasts.value.length > MAX_VISIBLE) {
      const oldest = toasts.value[0]
      if (oldest) dismiss(oldest.id)
    }
    timers.set(
      id,
      setTimeout(() => dismiss(id), DURATION_MS[kind]),
    )
    return id
  }

  return {
    toasts,
    dismiss,
    success: (title: string, description?: string) => push('success', title, description),
    error: (title: string, description?: string) => push('error', title, description),
    info: (title: string, description?: string) => push('info', title, description),
  }
})
