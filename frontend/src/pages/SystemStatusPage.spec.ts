import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import SystemStatusPage from './SystemStatusPage.vue'
import { healthService } from '@/services/healthService'
import { useToastStore } from '@/stores/toast'
import { ApiError } from '@/utils/apiError'
import type { HealthReport } from '@/types/api'

const upReport: HealthReport = {
  status: 'up',
  checks: { postgres: { status: 'up', latency_ms: 3 } },
  build: { version: '1.2.0', commit: 'abcdef1234567', build_time: '2026-11-17T10:00:00Z' },
  uptime: '2h15m0s',
  checked_at: '2026-11-17T12:00:00Z',
}

function mountPage() {
  return mount(SystemStatusPage)
}

describe('SystemStatusPage', () => {
  beforeEach(() => setActivePinia(createPinia()))
  afterEach(() => vi.useRealTimers())

  it('mostra skeleton enquanto carrega', () => {
    vi.spyOn(healthService, 'readiness').mockReturnValue(new Promise(() => {}))
    const w = mountPage()
    expect(w.find('[data-testid="status-skeleton"]').exists()).toBe(true)
  })

  it('exibe o relatório quando tudo está operacional', async () => {
    vi.spyOn(healthService, 'readiness').mockResolvedValue(upReport)
    const w = mountPage()
    await flushPromises()

    expect(w.text()).toContain('Todos os sistemas operacionais')
    expect(w.text()).toContain('Banco de dados (PostgreSQL)')
    expect(w.text()).toContain('3 ms')
    expect(w.text()).toContain('1.2.0 (abcdef1)')
    expect(w.text()).toContain('2h 15min')
  })

  it('sinaliza dependência fora do ar', async () => {
    vi.spyOn(healthService, 'readiness').mockResolvedValue({
      ...upReport,
      status: 'down',
      checks: { postgres: { status: 'down', latency_ms: 2000, error: 'indisponível' } },
    })
    const w = mountPage()
    await flushPromises()

    expect(w.text()).toContain('Instabilidade detectada')
    expect(w.text()).toContain('Offline')
    expect(w.text()).toContain('indisponível')
  })

  it('mostra estado vazio quando não há dependências', async () => {
    vi.spyOn(healthService, 'readiness').mockResolvedValue({ ...upReport, checks: {} })
    const w = mountPage()
    await flushPromises()

    expect(w.text()).toContain('Nenhuma dependência monitorada')
  })

  it('mostra erro com código de suporte e permite tentar novamente', async () => {
    const spy = vi
      .spyOn(healthService, 'readiness')
      .mockRejectedValueOnce(
        new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'Sem conexão', requestId: 'req-42' }),
      )
      .mockResolvedValueOnce(upReport)

    const w = mountPage()
    await flushPromises()

    expect(w.find('[role="alert"]').text()).toContain('Sem conexão')
    expect(w.text()).toContain('req-42')

    await w.find('[role="alert"] button').trigger('click')
    await flushPromises()

    expect(spy).toHaveBeenCalledTimes(2)
    expect(w.text()).toContain('Todos os sistemas operacionais')
  })

  it('atualização manual com falha mantém os dados e avisa por toast', async () => {
    vi.spyOn(healthService, 'readiness')
      .mockResolvedValueOnce(upReport)
      .mockRejectedValueOnce(new ApiError({ status: 503, code: 'HTTP_503', message: 'Fora do ar' }))

    const w = mountPage()
    await flushPromises()
    const atualizar = w.findAll('button').find((b) => b.text().includes('Atualizar'))!
    await atualizar.trigger('click')
    await flushPromises()

    expect(w.text()).toContain('Todos os sistemas operacionais')
    const toast = useToastStore()
    expect(toast.toasts.at(-1)?.title).toBe('Falha ao atualizar o status')
  })

  it('atualiza automaticamente a cada 30s e para ao desmontar', async () => {
    vi.useFakeTimers()
    const spy = vi.spyOn(healthService, 'readiness').mockResolvedValue(upReport)

    const w = mountPage()
    await flushPromises()
    expect(spy).toHaveBeenCalledTimes(1)

    await vi.advanceTimersByTimeAsync(30_000)
    expect(spy).toHaveBeenCalledTimes(2)

    w.unmount()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(spy).toHaveBeenCalledTimes(2)
  })
})
