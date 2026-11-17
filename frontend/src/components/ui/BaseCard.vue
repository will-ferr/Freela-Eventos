<script setup lang="ts">
defineProps<{ title?: string; subtitle?: string }>()
</script>

<template>
  <section class="rounded-xl border border-line bg-surface shadow-sm">
    <header
      v-if="title || $slots.actions"
      class="flex items-start justify-between gap-4 border-b border-line px-5 py-4"
    >
      <div>
        <h2 v-if="title" class="text-sm font-semibold">{{ title }}</h2>
        <p v-if="subtitle" class="mt-0.5 text-xs text-muted">{{ subtitle }}</p>
      </div>
      <slot name="actions" />
    </header>
    <div class="p-5">
      <slot />
    </div>
  </section>
</template>
