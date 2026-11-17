// Contratos da API. Espelham os tipos do backend (ver Swagger em /api/v1/docs).

export interface ApiFieldError {
  field: string
  message: string
}

export interface ApiErrorBody {
  error: {
    code: string
    message: string
    fields?: ApiFieldError[]
    request_id?: string
  }
}

export type HealthStatus = 'up' | 'down'

export interface HealthCheckResult {
  status: HealthStatus
  latency_ms: number
  error?: string
}

export interface BuildInfo {
  version: string
  commit: string
  build_time: string
}

export interface HealthReport {
  status: HealthStatus
  checks: Record<string, HealthCheckResult>
  build: BuildInfo
  uptime: string
  checked_at: string
}
