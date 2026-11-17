import { http } from './http'
import type { HealthReport } from '@/types/api'

export const healthService = {
  /**
   * O endpoint responde 503 com o relatório completo quando alguma dependência
   * está fora. Isso não é falha de comunicação: o relatório é devolvido normalmente.
   */
  async readiness(signal?: AbortSignal): Promise<HealthReport> {
    const res = await http.get<HealthReport>('/health/ready', {
      signal,
      validateStatus: (s) => s === 200 || s === 503,
    })
    return res.data
  },
}
