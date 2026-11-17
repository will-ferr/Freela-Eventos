# syntax=docker/dockerfile:1

# ---------- build ----------
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache de dependências separado do código-fonte.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown
ENV CGO_ENABLED=0 GOOS=linux

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" && \
    go build -trimpath -ldflags "$LDFLAGS" -o /out/api ./cmd/api && \
    go build -trimpath -ldflags "-s -w" -o /out/migrate ./cmd/migrate

# ---------- desenvolvimento (hot reload com air) ----------
FROM golang:1.26-alpine AS dev
WORKDIR /src
RUN go install github.com/air-verse/air@v1.61.7
COPY go.mod go.sum ./
RUN go mod download
CMD ["air", "-c", ".air.toml"]

# ---------- runtime ----------
# distroless: sem shell nem gerenciador de pacotes, roda como usuário não-root.
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 CMD ["/app/api", "-healthcheck"]
ENTRYPOINT ["/app/api"]
