import { describe, expect, it, vi } from 'vitest'
import { createApp, defineComponent, h } from 'vue'
import { installGlobalErrorHandlers } from './observability'
import { ApiError } from './apiError'

describe('installGlobalErrorHandlers', () => {
  it('registra erros de renderização com request id e notifica o usuário', () => {
    const log = vi.spyOn(console, 'error').mockImplementation(() => {})
    const notify = vi.fn()

    const Broken = defineComponent({
      setup() {
        throw new ApiError({ status: 500, code: 'INTERNAL_ERROR', message: 'x', requestId: 'r-7' })
      },
      render: () => h('div'),
    })
    const app = createApp(Broken)
    installGlobalErrorHandlers(app, notify)
    app.mount(document.createElement('div'))

    expect(notify).toHaveBeenCalledOnce()
    expect(log).toHaveBeenCalledWith(
      '[workevent]',
      expect.objectContaining({ source: 'vue', code: 'INTERNAL_ERROR', requestId: 'r-7' }),
    )
  })

  it('captura promises rejeitadas sem tratamento', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    const notify = vi.fn()
    installGlobalErrorHandlers(createApp({ render: () => null }), notify)

    const event = new Event('unhandledrejection') as Event & { reason?: unknown }
    event.reason = new Error('promessa solta')
    window.dispatchEvent(event)

    expect(notify).toHaveBeenCalled()
  })
})
