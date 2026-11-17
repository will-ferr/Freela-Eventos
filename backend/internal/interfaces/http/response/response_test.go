package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/will-ferr/Freela-Eventos/backend/internal/domain/shared"
)

func init() { gin.SetMode(gin.TestMode) }

func TestStatusFor(t *testing.T) {
	cases := map[error]int{
		shared.NewValidation("V", "v"):   http.StatusUnprocessableEntity,
		shared.NewNotFound("N", "n"):     http.StatusNotFound,
		shared.NewConflict("C", "c"):     http.StatusConflict,
		shared.NewUnauthorized("U", "u"): http.StatusUnauthorized,
		shared.NewForbidden("F", "f"):    http.StatusForbidden,
		shared.NewBusinessRule("B", "b"): http.StatusUnprocessableEntity,
		errors.New("qualquer"):           http.StatusInternalServerError,
	}
	for err, want := range cases {
		if got := StatusFor(err); got != want {
			t.Errorf("StatusFor(%v) = %d, esperado %d", err, got, want)
		}
	}
}

func run(err error) (*httptest.ResponseRecorder, ErrorBody) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Writer.Header().Set(HeaderRequestID, "req-1")
	Abort(c, err)

	var body ErrorBody
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w, body
}

func TestAbort_DomainErrorIsExposed(t *testing.T) {
	err := fmt.Errorf("usecase: %w", shared.NewValidation("VALIDATION_ERROR", "Dados inválidos.",
		shared.FieldError{Field: "email", Message: "obrigatório"}))

	w, body := run(err)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", w.Code)
	}
	if body.Error.Code != "VALIDATION_ERROR" || len(body.Error.Fields) != 1 {
		t.Errorf("corpo inesperado: %+v", body)
	}
	if body.Error.RequestID != "req-1" {
		t.Errorf("request_id deveria ser incluído para suporte: %+v", body)
	}
}

func TestAbort_InternalErrorIsHidden(t *testing.T) {
	w, body := run(errors.New(`pq: password authentication failed for user "admin"`))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", w.Code)
	}
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("code = %q", body.Error.Code)
	}
	if body.Error.Message == "" || strings.Contains(body.Error.Message, "password") {
		t.Errorf("detalhe interno vazou: %q", body.Error.Message)
	}
}
