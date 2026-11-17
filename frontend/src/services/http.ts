import axios, { type AxiosInstance } from 'axios'
import { toApiError } from '@/utils/apiError'

/**
 * Cliente HTTP único da aplicação.
 * - baseURL relativa: em dev o Vite faz proxy, em produção o Nginx.
 * - withCredentials: necessário para o cookie httpOnly de refresh (Etapa 2).
 * - Todo erro sai como ApiError (ver utils/apiError.ts).
 */
export function createHttpClient(baseURL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'): AxiosInstance {
  const client = axios.create({
    baseURL,
    timeout: 15_000,
    withCredentials: true,
    headers: { Accept: 'application/json' },
  })

  client.interceptors.response.use(
    (res) => res,
    (err: unknown) => Promise.reject(toApiError(err)),
  )

  return client
}

export const http = createHttpClient()
