# WorkEvent

Plataforma de gestão e contratação de freelancers para bares, restaurantes, casas noturnas,
festivais e eventos: publicação de vagas, candidaturas, escalas, controle de presença,
pagamentos e avaliações.

[![CI](https://github.com/will-ferr/Freela-Eventos/actions/workflows/ci.yml/badge.svg)](https://github.com/will-ferr/Freela-Eventos/actions/workflows/ci.yml)

| Camada | Tecnologias |
|---|---|
| Backend | Go · Gin · Clean Architecture · GORM · PostgreSQL · golang-migrate · Swagger |
| Frontend | Vue 3 · TypeScript · Vite · Tailwind CSS · Pinia · Vue Router · Axios |
| Infra | Docker · Docker Compose · Nginx · GitHub Actions |
| Observabilidade | slog (JSON) · Prometheus · Grafana · OpenTelemetry · Jaeger |
| Testes | `go test` · testes de integração com PostgreSQL · Vitest · Vue Test Utils |

## Arquitetura

```
 Navegador ──HTTPS──► frontend (Nginx) ──/api/*──► backend (Go/Gin) ──► PostgreSQL
                         │  SPA estática               │  ▲
                         │  X-Request-ID               │  └── migrate (one-shot, antes da API)
                                                       ├── /metrics ◄── Prometheus ◄── Grafana
                                                       └── OTLP ──────► Jaeger
```

- Frontend e API na **mesma origem**: o Nginx faz proxy de `/api`, sem CORS em produção.
- O backend segue **Clean Architecture** (`domain → application → infrastructure / interfaces`).
- Decisões registradas em [`docs/adr`](docs/adr):
  [arquitetura](docs/adr/0001-clean-architecture.md) ·
  [migrations](docs/adr/0002-migrations-sql.md) ·
  [tokens](docs/adr/0003-estrategia-de-tokens.md) ·
  [observabilidade](docs/adr/0004-observabilidade.md)

## Como executar

Pré-requisito: Docker Desktop (ou Docker Engine + Compose v2).

```bash
cp .env.example .env          # ajuste POSTGRES_PASSWORD
docker compose up -d --build
```

| Serviço | URL |
|---|---|
| Aplicação | http://localhost:8000 |
| API (direto) | http://localhost:8080/api/v1/health/ready |
| Swagger | http://localhost:8000/api/v1/docs/index.html |

> **Porta ocupada?** Se outro PostgreSQL já usa a 5432 (ou qualquer porta acima), troque
> `POSTGRES_PORT`, `BACKEND_PORT` ou `FRONTEND_PORT` no `.env`.

### Modo desenvolvimento (hot reload)

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Frontend em http://localhost:5173 (Vite + HMR); backend recompila a cada alteração (air).

### Com observabilidade

```bash
OTEL_TRACING_ENABLED=true docker compose --profile observability up -d --build
```

| Ferramenta | URL | |
|---|---|---|
| Grafana | http://localhost:3000 | dashboard **WorkEvent — API** já provisionado |
| Prometheus | http://localhost:9090 | métricas e [alertas](observability/prometheus/alerts.yml) |
| Jaeger | http://localhost:16686 | traces por requisição e por query SQL |

> No PowerShell, defina a variável antes: `$env:OTEL_TRACING_ENABLED="true"; docker compose --profile observability up -d --build`

**Investigando um erro:** a tela de erro mostra um *código de suporte* (o `X-Request-ID`).
Ele aparece no access log do Nginx, em todos os logs da API (`request_id`) e leva ao
`trace_id` correspondente no Jaeger.

## Testes

```bash
# Backend — unitários
cd backend && go test -cover ./...

# Backend — integração (PostgreSQL real; banco workevent_test criado no modo dev)
TEST_DATABASE_URL="postgres://workevent:SENHA@localhost:5432/workevent_test?sslmode=disable" \
  go test -tags=integration -count=1 -p 1 ./...

# Frontend
cd frontend && npm test            # ou npm run test:coverage
```

O CI ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)) executa em todo push/PR:
gofmt, Swagger atualizado, `go vet`, golangci-lint, testes com race detector, testes de
integração contra PostgreSQL, ESLint, type-check, Vitest com cobertura, build e build das
imagens Docker.

O que já é coberto:

- **Backend:** validação de configuração; mapeamento de erros de domínio para HTTP (sem vazar
  detalhes internos); request id (inclusive contra log injection); recovery de panic; métricas
  por rota; rate limit por IP; headers de segurança e CORS; readiness com timeout;
  reversibilidade das migrations; constraints do banco; spans SQL sem valores de parâmetros.
- **Frontend:** normalização de erros da API; `useAsync` (corrida e cancelamento); toasts;
  tema; página de status (loading, erro, vazio, atualização automática); layout responsivo;
  captura global de erros.

## Estrutura

```
backend/
  cmd/api, cmd/migrate        pontos de entrada
  config/                     configuração via ambiente (validada no boot)
  internal/domain/            entidades e regras de negócio (sem frameworks)
  internal/application/       casos de uso
  internal/infrastructure/    postgres, logger, telemetria
  internal/interfaces/http/   handlers, middlewares, rotas
  migrations/                 SQL versionado (embutido no binário)
  docs/                       Swagger gerado
frontend/src/
  components/ composables/ layouts/ pages/ router/ services/ stores/ types/ utils/
observability/                Prometheus, alertas e Grafana provisionados
docs/adr/                     decisões arquiteturais
```

## Roadmap

- [x] **Etapa 1 — Fundação:** estrutura, Docker, PostgreSQL, migrations, integração front/back,
  testes e observabilidade
- [ ] Etapa 2 — Autenticação (JWT + refresh, RBAC, perfis)
- [ ] Etapa 3 — Freelancers, contratantes, estabelecimentos, categorias e disponibilidade
- [ ] Etapa 4 — Oportunidades, busca e candidaturas
- [ ] Etapa 5 — Eventos, agenda, escalas e presença
- [ ] Etapa 6 — Financeiro e avaliações
- [ ] Etapa 7 — Dashboards e notificações
- [ ] Etapa 8 — Hardening, testes E2E e deploy
