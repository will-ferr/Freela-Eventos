import type { Component } from 'vue'
import { Activity } from 'lucide-vue-next'

export interface NavItem {
  label: string
  to: string
  icon: Component
}

/**
 * Itens da sidebar. Só entram aqui telas que já existem — nada de links mortos.
 * A partir da Etapa 2 cada item ganha `roles` e a lista é filtrada pelo perfil.
 */
export const navigation: NavItem[] = [{ label: 'Status do sistema', to: '/status', icon: Activity }]
