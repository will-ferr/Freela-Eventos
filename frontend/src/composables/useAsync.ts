import { getCurrentScope, onScopeDispose, ref, shallowRef, type Ref, type ShallowRef } from 'vue'
import { ApiError, toApiError } from '@/utils/apiError'

export interface UseAsyncReturn<T, A extends unknown[]> {
  data: ShallowRef<T | null>
  error: ShallowRef<ApiError | null>
  loading: Ref<boolean>
  /** Executa a operação; retorna o resultado ou null em caso de erro/cancelamento. */
  execute: (...args: A) => Promise<T | null>
}

/**
 * Padroniza os três estados de qualquer tela que busca dados: loading, erro e
 * sucesso. Também cancela a requisição anterior quando uma nova é disparada
 * (evita condição de corrida) e quando o componente é desmontado.
 */
export function useAsync<T, A extends unknown[] = []>(
  fn: (signal: AbortSignal, ...args: A) => Promise<T>,
  opts: { immediate?: boolean; keepDataOnError?: boolean } = {},
): UseAsyncReturn<T, A> {
  const data = shallowRef<T | null>(null)
  const error = shallowRef<ApiError | null>(null)
  const loading = ref(false)
  let controller: AbortController | null = null

  async function execute(...args: A): Promise<T | null> {
    controller?.abort()
    const current = new AbortController()
    controller = current

    loading.value = true
    error.value = null
    try {
      const result = await fn(current.signal, ...args)
      if (current.signal.aborted) return null
      data.value = result
      return result
    } catch (err) {
      if (current.signal.aborted) return null
      error.value = toApiError(err)
      if (!opts.keepDataOnError) data.value = null
      return null
    } finally {
      if (controller === current) loading.value = false
    }
  }

  if (getCurrentScope()) onScopeDispose(() => controller?.abort())

  if (opts.immediate) {
    void execute(...([] as unknown as A))
  }

  return { data, error, loading, execute }
}
