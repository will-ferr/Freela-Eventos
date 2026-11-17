<script setup lang="ts">
import { RouterView } from 'vue-router'
import ToastContainer from '@/components/ui/ToastContainer.vue'
import { useThemeStore } from '@/stores/theme'

useThemeStore() // aplica o tema salvo/do sistema na inicialização
</script>

<template>
  <RouterView />
  <ToastContainer />
</template>
