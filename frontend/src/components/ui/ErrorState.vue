<script setup lang="ts">
import { TriangleAlert, RotateCw } from 'lucide-vue-next'
import BaseButton from './BaseButton.vue'
import type { ApiError } from '@/utils/apiError'

defineProps<{ error: ApiError; retrying?: boolean }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div role="alert" class="flex flex-col items-center justify-center px-6 py-12 text-center">
    <div class="mb-4 flex size-12 items-center justify-center rounded-full bg-danger-soft text-danger">
      <TriangleAlert class="size-6" aria-hidden="true" />
    </div>
    <h3 class="text-sm font-semibold">Não foi possível carregar</h3>
    <p class="mt-1 max-w-sm text-sm text-muted">{{ error.message }}</p>
    <p v-if="error.requestId" class="mt-2 font-mono text-xs text-muted">
      Código de suporte: {{ error.requestId }}
    </p>
    <BaseButton class="mt-5" variant="secondary" :loading="retrying" @click="$emit('retry')">
      <RotateCw v-if="!retrying" class="size-4" aria-hidden="true" />
      Tentar novamente
    </BaseButton>
  </div>
</template>
