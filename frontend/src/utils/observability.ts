import type { App } from 'vue'
import { toApiError } from './apiError'

type Notify = (title: string, description?: string) => void

/**
 * Captura erros não tratados (renderização, promises soltas) e os registra de
 * forma estruturada. Ponto único para plugar um serviço como Sentry no futuro.
 */
export function installGlobalErrorHandlers(app: App, notify: Notify) {
  const report = (source: string, err: unknown) => {
    const apiErr = toApiError(err)
    console.error('[workevent]', {
      source,
      code: apiErr.code,
      message: apiErr.message,
      requestId: apiErr.requestId,
      route: window.location.pathname,
      error: err,
    })
    notify('Algo deu errado', 'Tente novamente. Se persistir, recarregue a página.')
  }

  app.config.errorHandler = (err) => report('vue', err)
  window.addEventListener('unhandledrejection', (e) => report('promise', e.reason))
}
