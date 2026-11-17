<script setup lang="ts">
import { CircleCheck, CircleAlert, Info, X } from 'lucide-vue-next'
import { useToastStore, type ToastKind } from '@/stores/toast'

const store = useToastStore()

const icons = { success: CircleCheck, error: CircleAlert, info: Info } as const
const tones: Record<ToastKind, string> = {
  success: 'text-success',
  error: 'text-danger',
  info: 'text-info',
}
</script>

<template>
  <!-- aria-live: leitores de tela anunciam novas notificações -->
  <div
    aria-live="polite"
    class="pointer-events-none fixed inset-x-0 bottom-0 z-50 flex flex-col items-center gap-2 p-4 sm:inset-x-auto sm:right-0 sm:items-end"
  >
    <TransitionGroup
      enter-from-class="translate-y-2 opacity-0"
      enter-active-class="transition duration-200"
      leave-to-class="opacity-0"
      leave-active-class="transition duration-150"
    >
      <div
        v-for="t in store.toasts"
        :key="t.id"
        :role="t.kind === 'error' ? 'alert' : 'status'"
        class="pointer-events-auto flex w-full max-w-sm items-start gap-3 rounded-xl border border-line bg-surface p-4 shadow-lg"
      >
        <component :is="icons[t.kind]" class="mt-0.5 size-5 shrink-0" :class="tones[t.kind]" />
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium">{{ t.title }}</p>
          <p v-if="t.description" class="mt-0.5 break-words text-sm text-muted">
            {{ t.description }}
          </p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-muted hover:bg-surface-2 hover:text-ink"
          aria-label="Fechar notificação"
          @click="store.dismiss(t.id)"
        >
          <X class="size-4" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>
