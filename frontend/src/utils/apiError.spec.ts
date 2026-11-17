import { describe, expect, it } from 'vitest'
import { AxiosError, AxiosHeaders, type AxiosResponse } from 'axios'
import { ApiError, toApiError } from './apiError'

function axiosErrorWith(status: number, data: unknown, headers: Record<string, string> = {}) {
  const config = { headers: new AxiosHeaders() }
  const response = {
    status,
    data,
    headers,
    statusText: '',
    config,
  } as unknown as AxiosResponse
  return new AxiosError('Request failed', 'ERR_BAD_RESPONSE', config, null, response)
}

describe('toApiError', () => {
  it('usa o corpo padronizado da API quando presente', () => {
    const err = toApiError(
      axiosErrorWith(422, {
        error: {
          code: 'VALIDATION_ERROR',
          message: 'Dados inválidos.',
          fields: [{ field: 'email', message: 'obrigatório' }],
          request_id: 'req-1',
        },
      }),
    )

    expect(err).toBeInstanceOf(ApiError)
    expect(err.status).toBe(422)
    expect(err.code).toBe('VALIDATION_ERROR')
    expect(err.fields).toHaveLength(1)
    expect(err.requestId).toBe('req-1')
  })

  it('recorre ao header X-Request-ID e mensagem padrão quando o corpo não é da API', () => {
    // ex.: 502 gerado pelo Nginx com HTML
    const err = toApiError(axiosErrorWith(503, '<html>', { 'x-request-id': 'nginx-9' }))

    expect(err.code).toBe('HTTP_503')
    expect(err.message).toBe('Serviço temporariamente indisponível.')
    expect(err.requestId).toBe('nginx-9')
  })

  it('identifica falha de rede (sem resposta)', () => {
    const err = toApiError(new AxiosError('Network Error', 'ERR_NETWORK'))

    expect(err.isNetworkError).toBe(true)
    expect(err.code).toBe('NETWORK_ERROR')
  })

  it('identifica timeout', () => {
    const err = toApiError(new AxiosError('timeout', 'ECONNABORTED'))
    expect(err.code).toBe('TIMEOUT')
  })

  it('converte erros genéricos e preserva ApiError existente', () => {
    expect(toApiError(new Error('boom')).message).toBe('boom')
    expect(toApiError('???').code).toBe('UNKNOWN')

    const original = new ApiError({ status: 409, code: 'DUP', message: 'x' })
    expect(toApiError(original)).toBe(original)
  })
})
