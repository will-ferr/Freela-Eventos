// Command migrate aplica ou reverte as migrations do banco.
//
//	migrate up            aplica todas as pendentes
//	migrate down [n]      reverte n migrations (padrão 1)
//	migrate version       mostra a versão atual
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"

	"github.com/will-ferr/Freela-Eventos/backend/config"
	"github.com/will-ferr/Freela-Eventos/backend/migrations"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("uso: migrate up | down [n] | version")
	}

	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}

	m, err := migrations.New(cfg.DB.URL())
	if err != nil {
		return err
	}
	defer m.Close()

	switch args[0] {
	case "up":
		err = m.Up()
	case "down":
		n := 1
		if len(args) > 1 {
			if n, err = strconv.Atoi(args[1]); err != nil || n < 1 {
				return fmt.Errorf("quantidade inválida: %q", args[1])
			}
		}
		err = m.Steps(-n)
	case "version":
		v, dirty, verr := m.Version()
		if errors.Is(verr, migrate.ErrNilVersion) {
			fmt.Println("nenhuma migration aplicada")
			return nil
		}
		if verr != nil {
			return verr
		}
		fmt.Printf("versão %d (dirty=%v)\n", v, dirty)
		return nil
	default:
		return fmt.Errorf("comando desconhecido %q", args[0])
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("nenhuma alteração pendente")
		return nil
	}
	if err != nil {
		return err
	}
	v, _, _ := m.Version()
	fmt.Printf("ok — versão atual: %d\n", v)
	return nil
}
