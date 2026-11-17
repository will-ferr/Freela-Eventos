import { describe, expect, it } from 'vitest'
import { effectScope } from 'vue'
import { useAsync } from './useAsync'
import { ApiError } from '@/utils/apiError'

function deferred<T>() {
  let resolve!: (v: T) => void
  let reject!: (e: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

describe('useAsync', () => {
  it('controla loading e armazena o resultado', async () => {
    const d = deferred<string>()
    const { data, loading, error, execute } = useAsync(() => d.promise)

    const running = execute()
    expect(loading.value).toBe(true)

    d.resolve('ok')
    await expect(running).resolves.toBe('ok')
    expect(loading.value).toBe(false)
    expect(data.value).toBe('ok')
    expect(error.value).toBeNull()
  })

  it('normaliza erros em ApiError', async () => {
    const { error, data, execute } = useAsync(() => Promise.reject(new Error('falhou')))

    await expect(execute()).resolves.toBeNull()
    expect(error.value).toBeInstanceOf(ApiError)
    expect(error.value?.message).toBe('falhou')
    expect(data.value).toBeNull()
  })

  it('mantém os dados anteriores em erro quando keepDataOnError', async () => {
    let fail = false
    const { data, error, execute } = useAsync(
      () => (fail ? Promise.reject(new Error('x')) : Promise.resolve(1)),
      { keepDataOnError: true },
    )

    await execute()
    fail = true
    await execute()

    expect(data.value).toBe(1)
    expect(error.value).not.toBeNull()
  })

  it('descarta resposta de requisição anterior (evita condição de corrida)', async () => {
    const first = deferred<string>()
    const second = deferred<string>()
    const calls = [first, second]
    const signals: AbortSignal[] = []
    const { data, loading, execute } = useAsync((signal) => {
      signals.push(signal)
      return calls.shift()!.promise
    })

    const p1 = execute()
    const p2 = execute()
    expect(signals[0]?.aborted).toBe(true)

    second.resolve('nova')
    first.resolve('antiga')
    await Promise.all([p1, p2])

    expect(data.value).toBe('nova')
    expect(loading.value).toBe(false)
  })

  it('cancela a requisição pendente quando o escopo é destruído', () => {
    let captured: AbortSignal | undefined
    const scope = effectScope()
    scope.run(() => {
      const { execute } = useAsync((signal) => {
        captured = signal
        return new Promise(() => {})
      })
      void execute()
    })

    scope.stop()
    expect(captured?.aborted).toBe(true)
  })
})
