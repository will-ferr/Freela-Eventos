import { beforeEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { defineComponent } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import AppLayout from './AppLayout.vue'

const Stub = defineComponent({ template: '<p>conteúdo</p>' })

function setup() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/',
        component: AppLayout,
        children: [
          { path: 'status', component: Stub, meta: { title: 'Status do sistema' } },
          { path: 'outra', component: Stub, meta: { title: 'Outra' } },
        ],
      },
    ],
  })
  return { router, pinia: createPinia() }
}

describe('AppLayout', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.classList.remove('dark')
  })

  it('exibe o título da rota e o conteúdo da página', async () => {
    const { router, pinia } = setup()
    await router.push('/status')
    const w = mount({ template: '<RouterView />' }, { global: { plugins: [router, pinia] } })
    await flushPromises()

    expect(w.find('header h1').text()).toBe('Status do sistema')
    expect(w.text()).toContain('conteúdo')
  })

  it('abre o menu mobile e fecha ao navegar', async () => {
    const { router, pinia } = setup()
    await router.push('/status')
    const w = mount({ template: '<RouterView />' }, { global: { plugins: [router, pinia] } })
    await flushPromises()

    const open = w.find('button[aria-label="Abrir menu"]')
    expect(open.attributes('aria-expanded')).toBe('false')
    await open.trigger('click')
    expect(w.find('#mobile-nav').exists()).toBe(true)

    await router.push('/outra')
    await flushPromises()
    expect(w.find('#mobile-nav').exists()).toBe(false)
  })

  it('alterna o tema', async () => {
    localStorage.setItem('workevent:theme', 'light')
    const { router, pinia } = setup()
    await router.push('/status')
    const w = mount({ template: '<RouterView />' }, { global: { plugins: [router, pinia] } })
    await flushPromises()

    await w.find('button[aria-label="Usar tema escuro"]').trigger('click')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(w.find('button[aria-label="Usar tema claro"]').exists()).toBe(true)
  })
})
