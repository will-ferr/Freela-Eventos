// Package response padroniza o formato de todas as respostas de erro da API e
// faz a tradução de erros de domínio para status HTTP.
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/will-ferr/Freela-Eventos/backend/internal/domain/shared"
)

const HeaderRequestID = "X-Request-ID"

// ErrorBody é o contrato de erro da API, documentado no Swagger.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code      string              `json:"code" example:"VALIDATION_ERROR"`
	Message   string              `json:"message" example:"Dados inválidos."`
	Fields    []shared.FieldError `json:"fields,omitempty"`
	RequestID string              `json:"request_id,omitempty" example:"0b6f0f4e-..."`
}

var statusByKind = map[shared.Kind]int{
	shared.KindValidation:   http.StatusUnprocessableEntity,
	shared.KindNotFound:     http.StatusNotFound,
	shared.KindConflict:     http.StatusConflict,
	shared.KindUnauthorized: http.StatusUnauthorized,
	shared.KindForbidden:    http.StatusForbidden,
	shared.KindBusinessRule: http.StatusUnprocessableEntity,
	shared.KindInternal:     http.StatusInternalServerError,
}

// StatusFor retorna o status HTTP correspondente ao erro.
func StatusFor(err error) int { return statusByKind[shared.KindOf(err)] }

// Body monta o corpo de erro. Erros internos nunca expõem a mensagem original.
func Body(c *gin.Context, err error) ErrorBody {
	detail := ErrorDetail{
		Code:      "INTERNAL_ERROR",
		Message:   "Ocorreu um erro inesperado. Tente novamente em instantes.",
		RequestID: c.Writer.Header().Get(HeaderRequestID),
	}
	if de, ok := asDomain(err); ok && de.Kind != shared.KindInternal {
		detail.Code = de.Code
		detail.Message = de.Message
		detail.Fields = de.Fields
	}
	return ErrorBody{Error: detail}
}

// Abort encerra a requisição com um erro de domínio.
func Abort(c *gin.Context, err error) {
	_ = c.Error(err)
	c.AbortWithStatusJSON(StatusFor(err), Body(c, err))
}

// AbortWith encerra a requisição com status e código explícitos (erros de
// transporte, como 404 de rota ou 429, que não pertencem ao domínio).
func AbortWith(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: ErrorDetail{
		Code:      code,
		Message:   message,
		RequestID: c.Writer.Header().Get(HeaderRequestID),
	}})
}

func asDomain(err error) (*shared.Error, bool) {
	var de *shared.Error
	ok := errors.As(err, &de)
	return de, ok
}
