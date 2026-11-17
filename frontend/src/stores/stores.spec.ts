import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useToastStore } from './toast'
import { useThemeStore } from './theme'

beforeEach(() => setActivePinia(createPinia()))

describe('toast store', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('adiciona e remove automaticamente após o tempo do tipo', () => {
    const store = useToastStore()
    store.success('Salvo')
    store.error('Falhou')

    expect(store.toasts).toHaveLength(2)

    vi.advanceTimersByTime(4000) // success expira
    expect(store.toasts.map((t) => t.title)).toEqual(['Falhou'])

    vi.advanceTimersByTime(4000) // error expira (8s)
    expect(store.toasts).toHaveLength(0)
  })

  it('limita a quantidade visível descartando o mais antigo', () => {
    const store = useToastStore()
    for (let i = 1; i <= 6; i++) store.info(`t${i}`)

    expect(store.toasts.map((t) => t.title)).toEqual(['t3', 't4', 't5', 't6'])
  })

  it('permite fechar manualmente', () => {
    const store = useToastStore()
    const id = store.info('x')
    store.dismiss(id)
    expect(store.toasts).toHaveLength(0)
  })
})

describe('theme store', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.classList.remove('dark')
  })

  it('usa a preferência salva e aplica a classe no <html>', () => {
    localStorage.setItem('workevent:theme', 'dark')
    const store = useThemeStore()

    expect(store.theme).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('alterna e persiste', () => {
    localStorage.setItem('workevent:theme', 'light')
    const store = useThemeStore()

    store.toggle()

    expect(store.theme).toBe('dark')
    expect(localStorage.getItem('workevent:theme')).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('continua funcionando se o localStorage estiver bloqueado', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('SecurityError')
    })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('SecurityError')
    })

    const store = useThemeStore()
    expect(() => store.toggle()).not.toThrow()
  })
})
