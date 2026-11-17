# ADR-0003 — Access token em memória, refresh token em cookie httpOnly

**Status:** aceito (implementação na Etapa 2) · **Data:** 2026-11-17

## Contexto
O app é uma SPA usada majoritariamente no celular. Guardar tokens em `localStorage` os expõe
a qualquer XSS.

## Decisão
- **Access token**: JWT de 15 minutos, devolvido no corpo do login e mantido **apenas em
  memória** (Pinia). Enviado no header `Authorization: Bearer`.
- **Refresh token**: valor opaco aleatório de 7 dias, em cookie `httpOnly; Secure;
  SameSite=Strict; Path=/api/v1/auth`. No banco fica apenas o **hash SHA-256**.
- **Rotação**: cada uso gera um novo refresh token. Se um token já usado reaparecer, toda a
  família é revogada (indício de roubo).
- Frontend e API são servidos na **mesma origem** (Nginx faz proxy de `/api`), então o
  cookie `SameSite=Strict` funciona e o CORS fica restrito ao desenvolvimento.

## Consequências
- Recarregar a página exige uma chamada a `/auth/refresh` para obter novo access token
  (transparente para o usuário).
- O interceptor do Axios enfileira requisições durante o refresh para não disparar
  vários refresh em paralelo.
