<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ tone: 'success' | 'danger' | 'info' | 'neutral'; label: string }>()

const classes = computed(
  () =>
    ({
      success: 'bg-success-soft text-success',
      danger: 'bg-danger-soft text-danger',
      info: 'bg-info-soft text-info',
      neutral: 'bg-surface-2 text-muted',
    })[props.tone],
)
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium"
    :class="classes"
  >
    <span class="size-1.5 rounded-full bg-current" aria-hidden="true" />
    {{ label }}
  </span>
</template>
