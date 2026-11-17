# syntax=docker/dockerfile:1

# ---------- dependências ----------
FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

# ---------- desenvolvimento (Vite com HMR) ----------
FROM deps AS dev
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev"]

# ---------- build ----------
FROM deps AS build
COPY . .
ARG APP_VERSION=dev
ENV VITE_APP_VERSION=${APP_VERSION}
RUN npm run build

# ---------- runtime ----------
# Imagem não-root; o template é processado por envsubst na inicialização,
# permitindo trocar o upstream da API sem rebuild.
FROM nginxinc/nginx-unprivileged:1.27-alpine AS runtime
ENV API_UPSTREAM=backend:8080
ENV NGINX_ENVSUBST_FILTER=^API_
COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY security-headers.conf /etc/nginx/snippets/security-headers.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1
