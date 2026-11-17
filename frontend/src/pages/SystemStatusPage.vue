<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount } from 'vue'
import { Database, RotateCw, Server, ServerOff } from 'lucide-vue-next'
import BaseCard from '@/components/ui/BaseCard.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { useAsync } from '@/composables/useAsync'
import { healthService } from '@/services/healthService'
import { useToastStore } from '@/stores/toast'
import { formatDateTime, formatGoDuration, formatLatency, shortCommit } from '@/utils/format'

const AUTO_REFRESH_MS = 30_000

const toast = useToastStore()
const { data: report, error, loading, execute } = useAsync(
  (signal) => healthService.readiness(signal),
  { keepDataOnError: true },
)

const checks = computed(() => Object.entries(report.value?.checks ?? {}))
const isUp = computed(() => report.value?.status === 'up')
const frontendVersion = import.meta.env.VITE_APP_VERSION ?? 'dev'

const checkLabels: Record<string, string> = { postgres: 'Banco de dados (PostgreSQL)' }

async function refresh(manual = false) {
  const result = await execute()
  if (!manual) return
  if (result) {
    toast.success('Status atualizado')
  } else if (error.value) {
    const id = error.value.requestId ? ` (código ${error.value.requestId})` : ''
    toast.error('Falha ao atualizar o status', error.value.message + id)
  }
}

// Atualização automática apenas com a aba visível, para não gastar recursos.
let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  void refresh()
  timer = setInterval(() => {
    if (document.visibilityState === 'visible' && !loading.value) void refresh()
  }, AUTO_REFRESH_MS)
})
onBeforeUnmount(() => clearInterval(timer))
</script>

<template>
  <div class="space-y-6">
    <p class="text-sm text-muted">
      Saúde da API e das dependências. Atualiza automaticamente a cada 30 segundos.
    </p>

    <!-- Carregamento inicial (inclui o primeiro frame, antes do onMounted) -->
    <div v-if="!report && !error" class="space-y-6" data-testid="status-skeleton">
      <SkeletonBlock class="h-28 w-full" />
      <SkeletonBlock class="h-40 w-full" />
    </div>

    <!-- Erro sem dados anteriores (ex.: API fora do ar) -->
    <BaseCard v-else-if="error && !report">
      <ErrorState :error="error" :retrying="loading" @retry="refresh(true)" />
    </BaseCard>

    <template v-else-if="report">
      <BaseCard>
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex items-center gap-4">
            <div
              class="flex size-12 shrink-0 items-center justify-center rounded-xl"
              :class="isUp ? 'bg-success-soft text-success' : 'bg-danger-soft text-danger'"
            >
              <Server v-if="isUp" class="size-6" aria-hidden="true" />
              <ServerOff v-else class="size-6" aria-hidden="true" />
            </div>
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-lg font-semibold">
                  {{ isUp ? 'Todos os sistemas operacionais' : 'Instabilidade detectada' }}
                </h2>
                <StatusBadge
                  :tone="isUp ? 'success' : 'danger'"
                  :label="isUp ? 'Operacional' : 'Degradado'"
                />
              </div>
              <p class="mt-1 text-sm text-muted">
                Verificado em {{ formatDateTime(report.checked_at) }} · no ar há
                {{ formatGoDuration(report.uptime) }}
              </p>
            </div>
          </div>
          <BaseButton variant="secondary" :loading="loading" @click="refresh(true)">
            <RotateCw v-if="!loading" class="size-4" aria-hidden="true" />
            Atualizar
          </BaseButton>
        </div>
      </BaseCard>

      <BaseCard title="Dependências" subtitle="Cada serviço externo verificado pela API">
        <EmptyState
          v-if="checks.length === 0"
          :icon="Database"
          title="Nenhuma dependência monitorada"
          description="A API não registrou verificações de saúde."
        />
        <ul v-else class="-my-3 divide-y divide-line">
          <li
            v-for="[name, check] in checks"
            :key="name"
            class="flex items-center justify-between gap-4 py-3"
          >
            <div class="flex min-w-0 items-center gap-3">
              <Database class="size-5 shrink-0 text-muted" aria-hidden="true" />
              <div class="min-w-0">
                <p class="truncate text-sm font-medium">{{ checkLabels[name] ?? name }}</p>
                <p v-if="check.error" class="text-xs text-danger">{{ check.error }}</p>
              </div>
            </div>
            <div class="flex shrink-0 items-center gap-3">
              <span class="hidden font-mono text-xs text-muted sm:inline">
                {{ formatLatency(check.latency_ms) }}
              </span>
              <StatusBadge
                :tone="check.status === 'up' ? 'success' : 'danger'"
                :label="check.status === 'up' ? 'Online' : 'Offline'"
              />
            </div>
          </li>
        </ul>
      </BaseCard>

      <BaseCard title="Versões">
        <dl class="grid grid-cols-1 gap-4 text-sm sm:grid-cols-3">
          <div>
            <dt class="text-muted">API</dt>
            <dd class="mt-0.5 font-mono">
              {{ report.build.version }} ({{ shortCommit(report.build.commit) }})
            </dd>
          </div>
          <div>
            <dt class="text-muted">Build da API</dt>
            <dd class="mt-0.5 font-mono">{{ report.build.build_time }}</dd>
          </div>
          <div>
            <dt class="text-muted">Frontend</dt>
            <dd class="mt-0.5 font-mono">{{ frontendVersion }}</dd>
          </div>
        </dl>
      </BaseCard>
    </template>
  </div>
</template>
