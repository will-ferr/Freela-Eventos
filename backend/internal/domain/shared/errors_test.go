package shared

import (
	"errors"
	"fmt"
	"testing"
)

func TestKindOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want Kind
	}{
		{"validation", NewValidation("INVALID", "x"), KindValidation},
		{"not found", NewNotFound("NOT_FOUND", "x"), KindNotFound},
		{"wrapped by fmt", fmt.Errorf("camada: %w", NewConflict("DUP", "x")), KindConflict},
		{"unknown error", errors.New("boom"), KindInternal},
		{"nil-safe business", NewBusinessRule("RULE", "x"), KindBusinessRule},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := KindOf(tc.err); got != tc.want {
				t.Errorf("KindOf() = %v, esperado %v", got, tc.want)
			}
		})
	}
}

func TestWrap_PreservesCauseWithoutMutatingOriginal(t *testing.T) {
	base := NewNotFound("USER_NOT_FOUND", "usuário não encontrado")
	cause := errors.New("record not found")

	wrapped := base.Wrap(cause)

	if !errors.Is(wrapped, cause) {
		t.Error("errors.Is deveria encontrar a causa")
	}
	if base.Unwrap() != nil {
		t.Error("Wrap não pode alterar o erro original (sentinela compartilhada)")
	}
	if wrapped.Code != base.Code {
		t.Error("Wrap deve preservar o código")
	}
}
