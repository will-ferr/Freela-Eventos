import { afterEach, describe, expect, it } from 'vitest'
import type { AxiosAdapter, InternalAxiosRequestConfig } from 'axios'
import { AxiosError } from 'axios'
import { http } from './http'
import { healthService } from './healthService'
import { ApiError } from '@/utils/apiError'

const originalAdapter = http.defaults.adapter
afterEach(() => {
  http.defaults.adapter = originalAdapter
})

function respondWith(status: number, data: unknown): AxiosAdapter {
  return async (config: InternalAxiosRequestConfig) => {
    const response = { status, data, headers: {}, statusText: '', config }
    if (config.validateStatus && !config.validateStatus(status)) {
      throw new AxiosError('fail', 'ERR_BAD_RESPONSE', config, null, response)
    }
    return response
  }
}

const report = {
  status: 'down',
  checks: { postgres: { status: 'down', latency_ms: 2000, error: 'indisponível' } },
  build: { version: 'dev', commit: 'none', build_time: 'unknown' },
  uptime: '1m0s',
  checked_at: '2026-11-17T12:00:00Z',
}

describe('healthService.readiness', () => {
  it('devolve o relatório mesmo com 503 (dependência fora não é erro de comunicação)', async () => {
    http.defaults.adapter = respondWith(503, report)
    await expect(healthService.readiness()).resolves.toEqual(report)
  })

  it('converte outros status em ApiError', async () => {
    http.defaults.adapter = respondWith(500, {
      error: { code: 'INTERNAL_ERROR', message: 'Erro', request_id: 'r1' },
    })

    const err = await healthService.readiness().catch((e: unknown) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect((err as ApiError).requestId).toBe('r1')
  })
})
