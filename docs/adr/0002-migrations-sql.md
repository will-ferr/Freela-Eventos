# ADR-0002 — Migrations em SQL versionado (golang-migrate), sem AutoMigrate

**Status:** aceito · **Data:** 2026-11-17

## Contexto
Várias regras de integridade só existem de forma confiável no banco:
- exclusion constraint (`btree_gist`) para impedir escalas sobrepostas do mesmo freelancer;
- índices únicos parciais (`WHERE deleted_at IS NULL`) para exclusão lógica;
- `CHECK` de status e formatos;
- triggers de `updated_at`.

O `AutoMigrate` do GORM não expressa a maioria disso e não tem `down`.

## Decisão
- Migrations em SQL puro, numeradas (`000001_*.up.sql` / `.down.sql`), em `backend/migrations`.
- Embutidas no binário via `embed.FS`; executadas pelo comando `cmd/migrate`, que no Docker
  Compose roda como serviço one-shot **antes** da API subir.
- O GORM é usado só para consultas, nunca para alterar o schema.
- Toda migration tem `down`, validado em teste de integração (up → down → up).

## Consequências
- O schema é revisável em PR como qualquer código.
- Exige disciplina: toda alteração de model precisa da migration correspondente.
