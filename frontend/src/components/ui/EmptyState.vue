<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{ title: string; description?: string; icon?: Component }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center px-6 py-12 text-center">
    <div
      v-if="icon"
      class="mb-4 flex size-12 items-center justify-center rounded-full bg-surface-2 text-muted"
    >
      <component :is="icon" class="size-6" aria-hidden="true" />
    </div>
    <h3 class="text-sm font-semibold">{{ title }}</h3>
    <p v-if="description" class="mt-1 max-w-sm text-sm text-muted">{{ description }}</p>
    <div v-if="$slots.default" class="mt-5">
      <slot />
    </div>
  </div>
</template>
