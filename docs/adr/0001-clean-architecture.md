# ADR-0001 — Clean Architecture no backend

**Status:** aceito · **Data:** 2026-11-17

## Contexto
O domínio (contratações, escalas, conflitos de horário, pagamentos) tem regras de negócio
não triviais, que precisam ser testadas sem banco nem HTTP e sobreviver a trocas de
infraestrutura.

## Decisão
Quatro camadas em `backend/internal`, com dependências apontando sempre para dentro:

| Camada | Pode importar | Não pode importar |
|---|---|---|
| `domain` | stdlib | Gin, GORM, qualquer framework |
| `application` | `domain`, stdlib (`log/slog`) | HTTP, GORM |
| `infrastructure` | `domain`, `application`, libs externas | `interfaces` |
| `interfaces/http` | `application`, `domain` | acesso direto ao banco |

- Repositórios são **interfaces no domínio**, implementadas em `infrastructure/persistence`.
- Erros de domínio (`shared.Error`) carregam um `Kind` neutro; só `interfaces/http/response`
  conhece status HTTP.
- A composição (injeção de dependências) é manual, em `cmd/api/main.go`: explícita e sem
  "mágica" de container de DI.

## Consequências
- Casos de uso testáveis com fakes, em milissegundos.
- Mais arquivos e alguma duplicação (model GORM × entidade de domínio), aceitos em troca de
  isolamento.
