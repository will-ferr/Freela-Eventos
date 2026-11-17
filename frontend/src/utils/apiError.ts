import { isAxiosError } from 'axios'
import type { ApiErrorBody, ApiFieldError } from '@/types/api'

/**
 * Erro normalizado da aplicação. Toda falha de rede/API vira um ApiError, então
 * componentes nunca precisam conhecer o formato do Axios.
 */
export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly fields: ApiFieldError[]
  /** Id de correlação com os logs do backend — exibido ao usuário para suporte. */
  readonly requestId?: string

  constructor(opts: {
    status: number
    code: string
    message: string
    fields?: ApiFieldError[]
    requestId?: string
  }) {
    super(opts.message)
    this.name = 'ApiError'
    this.status = opts.status
    this.code = opts.code
    this.fields = opts.fields ?? []
    this.requestId = opts.requestId
  }

  get isNetworkError(): boolean {
    return this.status === 0
  }
}

function isApiErrorBody(data: unknown): data is ApiErrorBody {
  if (typeof data !== 'object' || data === null || !('error' in data)) return false
  const err = (data as { error: unknown }).error
  return typeof err === 'object' && err !== null && 'code' in err && 'message' in err
}

const fallbackByStatus: Record<number, string> = {
  401: 'Sua sessão expirou. Entre novamente.',
  403: 'Você não tem permissão para esta ação.',
  404: 'Recurso não encontrado.',
  429: 'Muitas requisições. Aguarde um instante.',
  503: 'Serviço temporariamente indisponível.',
}

export function toApiError(err: unknown): ApiError {
  if (err instanceof ApiError) return err

  if (isAxiosError(err)) {
    const res = err.response
    if (!res) {
      return new ApiError({
        status: 0,
        code: err.code === 'ECONNABORTED' ? 'TIMEOUT' : 'NETWORK_ERROR',
        message:
          err.code === 'ECONNABORTED'
            ? 'O servidor demorou para responder.'
            : 'Não foi possível conectar ao servidor. Verifique sua conexão.',
      })
    }

    const headerId = res.headers?.['x-request-id'] as string | undefined
    if (isApiErrorBody(res.data)) {
      const { code, message, fields, request_id } = res.data.error
      return new ApiError({
        status: res.status,
        code,
        message,
        fields,
        requestId: request_id ?? headerId,
      })
    }
    return new ApiError({
      status: res.status,
      code: `HTTP_${res.status}`,
      message: fallbackByStatus[res.status] ?? 'Ocorreu um erro inesperado.',
      requestId: headerId,
    })
  }

  return new ApiError({
    status: -1,
    code: 'UNKNOWN',
    message: err instanceof Error ? err.message : 'Ocorreu um erro inesperado.',
  })
}
