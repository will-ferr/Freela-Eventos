<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { navigation } from '@/router/navigation'

defineEmits<{ navigate: [] }>()
</script>

<template>
  <nav aria-label="Navegação principal" class="flex flex-1 flex-col gap-1 px-3">
    <RouterLink
      v-for="item in navigation"
      :key="item.to"
      :to="item.to"
      class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-sidebar-ink transition-colors hover:bg-sidebar-active hover:text-white"
      active-class="bg-sidebar-active !text-white"
      @click="$emit('navigate')"
    >
      <component :is="item.icon" class="size-4.5 shrink-0" aria-hidden="true" />
      {{ item.label }}
    </RouterLink>
  </nav>
</template>
