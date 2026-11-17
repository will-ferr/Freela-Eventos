<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { MapPinOff } from 'lucide-vue-next'
</script>

<template>
  <div class="flex min-h-dvh flex-col items-center justify-center px-6 text-center">
    <div class="mb-6 flex size-14 items-center justify-center rounded-full bg-brand-soft text-brand">
      <MapPinOff class="size-7" aria-hidden="true" />
    </div>
    <p class="text-sm font-semibold text-brand">Erro 404</p>
    <h1 class="mt-2 text-2xl font-semibold tracking-tight">Página não encontrada</h1>
    <p class="mt-2 max-w-sm text-sm text-muted">
      O endereço acessado não existe ou foi movido.
    </p>
    <RouterLink
      to="/"
      class="mt-6 inline-flex h-10 items-center rounded-lg bg-brand px-4 text-sm font-medium text-brand-ink hover:brightness-110"
    >
      Voltar ao início
    </RouterLink>
  </div>
</template>
