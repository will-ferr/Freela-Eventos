// Package handler contém os handlers HTTP. Handlers apenas traduzem HTTP para
// chamadas de caso de uso e o resultado de volta para JSON.
package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/will-ferr/Freela-Eventos/backend/internal/application/health"
)

// ReadinessChecker é a porta que o handler precisa; facilita testes.
type ReadinessChecker interface {
	Readiness(ctx context.Context) health.Report
}

type HealthHandler struct{ svc ReadinessChecker }

func NewHealthHandler(svc ReadinessChecker) *HealthHandler { return &HealthHandler{svc: svc} }

type LiveResponse struct {
	Status string `json:"status" example:"ok"`
}

// Live indica apenas que o processo está de pé (não consulta dependências),
// para que o orquestrador não reinicie o container por falha do banco.
//
//	@Summary	Liveness probe
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	LiveResponse
//	@Router		/health/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, LiveResponse{Status: "ok"})
}

// Ready verifica as dependências; retorna 503 quando alguma está indisponível,
// tirando a instância do balanceamento até se recuperar.
//
//	@Summary	Readiness probe
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	health.Report
//	@Failure	503	{object}	health.Report
//	@Router		/health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	report := h.svc.Readiness(c.Request.Context())
	status := http.StatusOK
	if report.Status != health.StatusUp {
		status = http.StatusServiceUnavailable
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(status, report)
}
