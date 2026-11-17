<script setup lang="ts">
import { computed } from 'vue'
import { LoaderCircle } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
    size?: 'sm' | 'md'
    type?: 'button' | 'submit'
    loading?: boolean
    disabled?: boolean
  }>(),
  { variant: 'primary', size: 'md', type: 'button', loading: false, disabled: false },
)

const classes = computed(() => [
  'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors',
  'disabled:cursor-not-allowed disabled:opacity-60',
  props.size === 'sm' ? 'h-8 px-3 text-sm' : 'h-10 px-4 text-sm',
  {
    primary: 'bg-brand text-brand-ink hover:brightness-110',
    secondary: 'border border-line bg-surface text-ink hover:bg-surface-2',
    ghost: 'text-muted hover:bg-surface-2 hover:text-ink',
    danger: 'bg-danger text-white hover:brightness-110',
  }[props.variant],
])
</script>

<template>
  <button :type="type" :class="classes" :disabled="disabled || loading" :aria-busy="loading">
    <LoaderCircle v-if="loading" class="size-4 animate-spin" aria-hidden="true" />
    <slot />
  </button>
</template>
