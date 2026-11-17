<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { Menu, X, Sun, Moon } from 'lucide-vue-next'
import SidebarNav from '@/components/navigation/SidebarNav.vue'
import BrandLogo from '@/components/navigation/BrandLogo.vue'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()
const route = useRoute()
const drawerOpen = ref(false)

// Fecha o menu mobile ao trocar de rota (inclusive via botão voltar).
watch(
  () => route.fullPath,
  () => (drawerOpen.value = false),
)
</script>

<template>
  <div class="min-h-dvh lg:pl-64">
    <!-- Sidebar fixa (desktop) -->
    <aside class="fixed inset-y-0 left-0 hidden w-64 flex-col bg-sidebar lg:flex">
      <div class="flex h-16 items-center px-6"><BrandLogo /></div>
      <SidebarNav />
    </aside>

    <!-- Drawer (mobile) -->
    <Transition
      enter-from-class="opacity-0"
      enter-active-class="transition-opacity duration-200"
      leave-to-class="opacity-0"
      leave-active-class="transition-opacity duration-150"
    >
      <div v-if="drawerOpen" class="fixed inset-0 z-40 lg:hidden">
        <div class="absolute inset-0 bg-black/50" aria-hidden="true" @click="drawerOpen = false" />
        <aside
          id="mobile-nav"
          class="relative flex h-full w-72 max-w-[85vw] flex-col bg-sidebar"
          aria-label="Menu"
        >
          <div class="flex h-16 items-center justify-between px-6">
            <BrandLogo />
            <button
              type="button"
              class="rounded-lg p-2 text-sidebar-ink hover:bg-sidebar-active"
              aria-label="Fechar menu"
              @click="drawerOpen = false"
            >
              <X class="size-5" />
            </button>
          </div>
          <SidebarNav @navigate="drawerOpen = false" />
        </aside>
      </div>
    </Transition>

    <!-- Topbar -->
    <header
      class="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-line bg-surface/90 px-4 backdrop-blur sm:px-6"
    >
      <button
        type="button"
        class="rounded-lg p-2 text-muted hover:bg-surface-2 lg:hidden"
        aria-label="Abrir menu"
        aria-controls="mobile-nav"
        :aria-expanded="drawerOpen"
        @click="drawerOpen = true"
      >
        <Menu class="size-5" />
      </button>
      <h1 class="truncate text-base font-semibold">{{ route.meta.title }}</h1>
      <div class="ml-auto flex items-center gap-1">
        <button
          type="button"
          class="rounded-lg p-2 text-muted hover:bg-surface-2 hover:text-ink"
          :aria-label="theme.theme === 'dark' ? 'Usar tema claro' : 'Usar tema escuro'"
          @click="theme.toggle()"
        >
          <Sun v-if="theme.theme === 'dark'" class="size-5" />
          <Moon v-else class="size-5" />
        </button>
      </div>
    </header>

    <main class="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6 lg:py-8">
      <RouterView />
    </main>
  </div>
</template>
