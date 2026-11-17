# Atalhos de desenvolvimento. No Windows sem make, use os comandos equivalentes do README.
.PHONY: help up up-obs dev down logs migrate test test-backend test-integration test-frontend lint swagger

COMPOSE     = docker compose
COMPOSE_DEV = docker compose -f docker-compose.yml -f docker-compose.dev.yml
TEST_DB_URL ?= postgres://workevent:$(or $(POSTGRES_PASSWORD),troque-esta-senha)@localhost:5432/workevent_test?sslmode=disable

help: ## Lista os comandos
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-18s %s\n", $$1, $$2}'

up: ## Sobe a stack (build de produção)
	$(COMPOSE) up -d --build

up-obs: ## Sobe a stack + Prometheus, Grafana e Jaeger
	OTEL_TRACING_ENABLED=true $(COMPOSE) --profile observability up -d --build

dev: ## Sobe em modo desenvolvimento (hot reload)
	$(COMPOSE_DEV) up --build

down: ## Para todos os containers
	$(COMPOSE) --profile observability down

logs: ## Acompanha os logs da API
	$(COMPOSE) logs -f backend

migrate: ## Aplica migrations pendentes
	$(COMPOSE) run --rm migrate up

test: test-backend test-frontend ## Roda todos os testes unitários

test-backend: ## Testes unitários do backend
	cd backend && go test -cover ./...

test-integration: ## Testes de integração (requer postgres do modo dev)
	cd backend && TEST_DATABASE_URL="$(TEST_DB_URL)" go test -tags=integration -count=1 -p 1 ./...

test-frontend: ## Testes do frontend
	cd frontend && npm test

lint: ## Lint do backend e do frontend
	cd backend && go vet ./... && golangci-lint run
	cd frontend && npm run lint && npm run typecheck

swagger: ## Regenera a documentação Swagger
	cd backend && go tool swag init -g cmd/api/main.go -o docs --parseInternal --outputTypes go,json
