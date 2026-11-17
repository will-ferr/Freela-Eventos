# ADR-0004 — Observabilidade: logs, métricas e traces correlacionados

**Status:** aceito · **Data:** 2026-11-17

## Contexto
Em produção, a pergunta "por que esta requisição falhou?" precisa ser respondida em minutos,
partindo do que o usuário vê na tela.

## Decisão
**Correlação ponta a ponta por `X-Request-ID`:**
Nginx gera (ou repassa) → API reaproveita se for seguro → aparece em todo log da API → volta
no header e no corpo de erro → o frontend exibe como *"Código de suporte"*.

| Sinal | Ferramenta | Detalhes |
|---|---|---|
| Logs | `log/slog` em JSON | `request_id`, `trace_id` e `span_id` automáticos via contexto; uma linha por requisição; queries lentas (> `DB_SLOW_QUERY_THRESHOLD`) e erros de SQL; probes em nível debug |
| Métricas | Prometheus (`/metrics`) | RED por rota (template, não path, para limitar cardinalidade), requisições em andamento, pool do banco, panics, rate limit, runtime Go, `build_info` |
| Traces | OpenTelemetry → OTLP/HTTP → Jaeger | span por requisição (otelgin) e por query (plugin GORM próprio, sem valores de parâmetros); propagação W3C `traceparent`; amostragem configurável |
| Saúde | `/health/live` e `/health/ready` | *live* não consulta dependências (evita restart em cascata); *ready* devolve 503 com o relatório |
| Alertas | regras do Prometheus | API fora, erro 5xx > 5%, p95 > 500 ms, panics, pool > 90%, pico de rate limit |
| Dashboards | Grafana provisionado | painel "WorkEvent — API" pronto ao subir o profile |

**Segurança:** `/metrics` fica fora de `/api` e o Nginx o bloqueia; erros internos nunca são
expostos ao cliente (só o código e o request id).

**Rate limit:** em memória por IP. Com múltiplas réplicas, migrar para Redis.

## Consequências
- Tracing desligado por padrão (`OTEL_TRACING_ENABLED=false`): custo zero quando não há coletor.
- A stack de observabilidade é opcional (`--profile observability`) para não pesar no dia a dia.
