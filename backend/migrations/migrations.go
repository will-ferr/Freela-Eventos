// Package migrations embute os arquivos SQL no binário, de modo que a imagem
// Docker não dependa de arquivos externos e os testes de integração usem
// exatamente as mesmas migrations da produção.
package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var FS embed.FS

// New cria uma instância do migrate apontando para o banco informado.
func New(databaseURL string) (*migrate.Migrate, error) {
	src, err := iofs.New(FS, ".")
	if err != nil {
		return nil, fmt.Errorf("lendo migrations embutidas: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("conectando para migrar: %w", err)
	}
	return m, nil
}

// Up aplica todas as migrations pendentes. Não ter nada a aplicar não é erro.
func Up(databaseURL string) error {
	m, err := New(databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
