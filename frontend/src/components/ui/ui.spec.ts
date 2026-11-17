import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import BaseButton from './BaseButton.vue'
import ErrorState from './ErrorState.vue'
import EmptyState from './EmptyState.vue'
import ToastContainer from './ToastContainer.vue'
import { useToastStore } from '@/stores/toast'
import { ApiError } from '@/utils/apiError'

describe('BaseButton', () => {
  it('fica desabilitado e sinaliza carregamento', () => {
    const w = mount(BaseButton, { props: { loading: true }, slots: { default: 'Salvar' } })
    const btn = w.find('button')

    expect(btn.attributes('disabled')).toBeDefined()
    expect(btn.attributes('aria-busy')).toBe('true')
    expect(btn.attributes('type')).toBe('button')
  })
})

describe('ErrorState', () => {
  it('mostra mensagem, código de suporte e emite retry', async () => {
    const error = new ApiError({ status: 500, code: 'X', message: 'Falhou', requestId: 'abc' })
    const w = mount(ErrorState, { props: { error } })

    expect(w.text()).toContain('Falhou')
    expect(w.text()).toContain('abc')

    await w.find('button').trigger('click')
    expect(w.emitted('retry')).toHaveLength(1)
  })

  it('omite o código quando não há request id', () => {
    const w = mount(ErrorState, {
      props: { error: new ApiError({ status: 0, code: 'N', message: 'Sem rede' }) },
    })
    expect(w.text()).not.toContain('Código de suporte')
  })
})

describe('EmptyState', () => {
  it('renderiza título, descrição e ação', () => {
    const w = mount(EmptyState, {
      props: { title: 'Nada aqui', description: 'Crie o primeiro item' },
      slots: { default: '<button>Criar</button>' },
    })
    expect(w.text()).toContain('Nada aqui')
    expect(w.text()).toContain('Crie o primeiro item')
    expect(w.find('button').text()).toBe('Criar')
  })
})

describe('ToastContainer', () => {
  it('renderiza toasts do store e permite fechar', async () => {
    setActivePinia(createPinia())
    const store = useToastStore()
    store.error('Erro ao salvar', 'Tente de novo')

    const w = mount(ToastContainer)
    expect(w.find('[role="alert"]').text()).toContain('Erro ao salvar')

    await w.find('button[aria-label="Fechar notificação"]').trigger('click')
    expect(store.toasts).toHaveLength(0)
  })
})
