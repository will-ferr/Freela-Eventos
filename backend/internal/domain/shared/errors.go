// Package shared contém tipos de domínio reutilizados por todos os módulos.
// Não pode depender de nenhum framework (HTTP, ORM etc.).
package shared

import (
	"errors"
	"fmt"
)

// Kind classifica o erro de forma independente de transporte. A camada HTTP
// traduz cada Kind para um status code; o domínio nunca conhece status codes.
type Kind int

const (
	KindInternal Kind = iota
	KindValidation
	KindNotFound
	KindConflict
	KindUnauthorized
	KindForbidden
	KindBusinessRule
)

func (k Kind) String() string {
	switch k {
	case KindValidation:
		return "validation"
	case KindNotFound:
		return "not_found"
	case KindConflict:
		return "conflict"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindBusinessRule:
		return "business_rule"
	default:
		return "internal"
	}
}

// FieldError descreve um problema em um campo específico da entrada.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error é o erro de domínio padrão. Code é estável e pode ser usado pelo
// frontend para decisões (ex.: "SCHEDULE_CONFLICT"); Message é legível.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Fields  []FieldError
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// Wrap anexa a causa original sem expô-la ao cliente.
func (e *Error) Wrap(cause error) *Error {
	cp := *e
	cp.cause = cause
	return &cp
}

func newErr(kind Kind, code, msg string) *Error {
	return &Error{Kind: kind, Code: code, Message: msg}
}

func NewValidation(code, msg string, fields ...FieldError) *Error {
	e := newErr(KindValidation, code, msg)
	e.Fields = fields
	return e
}
func NewNotFound(code, msg string) *Error     { return newErr(KindNotFound, code, msg) }
func NewConflict(code, msg string) *Error     { return newErr(KindConflict, code, msg) }
func NewUnauthorized(code, msg string) *Error { return newErr(KindUnauthorized, code, msg) }
func NewForbidden(code, msg string) *Error    { return newErr(KindForbidden, code, msg) }
func NewBusinessRule(code, msg string) *Error { return newErr(KindBusinessRule, code, msg) }

// KindOf retorna o Kind de qualquer erro; erros desconhecidos são internos.
func KindOf(err error) Kind {
	var de *Error
	if errors.As(err, &de) {
		return de.Kind
	}
	return KindInternal
}
