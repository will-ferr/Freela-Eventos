package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// reader acumula erros de conversão para que todos sejam reportados de uma vez,
// em vez de falhar na primeira variável inválida.
type reader struct {
	lookup func(string) (string, bool)
	errs   []error
}

func (r *reader) raw(key string) (string, bool) {
	v, ok := r.lookup(key)
	if !ok {
		return "", false
	}
	v = strings.TrimSpace(v)
	return v, v != ""
}

func (r *reader) str(key, def string) string {
	if v, ok := r.raw(key); ok {
		return v
	}
	return def
}

func (r *reader) int(key string, def int) int {
	v, ok := r.raw(key)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s deve ser um inteiro (recebido %q)", key, v))
		return def
	}
	return n
}

func (r *reader) float(key string, def float64) float64 {
	v, ok := r.raw(key)
	if !ok {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s deve ser numérico (recebido %q)", key, v))
		return def
	}
	return f
}

func (r *reader) bool(key string, def bool) bool {
	v, ok := r.raw(key)
	if !ok {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s deve ser true ou false (recebido %q)", key, v))
		return def
	}
	return b
}

func (r *reader) duration(key string, def time.Duration) time.Duration {
	v, ok := r.raw(key)
	if !ok {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s deve ser uma duração como 10s ou 5m (recebido %q)", key, v))
		return def
	}
	return d
}

func (r *reader) list(key string, def []string) []string {
	if v, ok := r.raw(key); ok {
		return splitList(v)
	}
	return def
}
