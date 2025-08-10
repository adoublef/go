package postgres_test

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	. "go.adoublef.dev/database/sql/postgres"
	"go.adoublef.dev/runtime/container/postgres"
	"go.adoublef.dev/testing/is"
)

//go:embed all:*.sql
var embedFS embed.FS

func TestFS(t *testing.T) {
	ctx := t.Context()

	p, err := container.ConnectionPool(ctx)
	is.OK(t, err) // container.ConnectionPool
	t.Cleanup(func() { p.Close() })

	fsys := &FS{
		URL: p.Config().ConnString(),
		FS:  embedFS,
	}

	is.OK(t, fsys.Up(ctx))   // fsys.Up
	is.OK(t, fsys.Down(ctx)) // fsys.Down
}

func TestMain(m *testing.M) {
	err := setup(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	err = cleanup(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

var container *postgres.Container

// setup initialises containers within the pacakge.
func setup(ctx context.Context) (err error) {
	container, err = postgres.Run(ctx, "")
	if err != nil {
		return
	}
	return
}

// cleanup stops all running containers for the pacakge.
func cleanup(ctx context.Context) (err error) {
	var cc = []testcontainers.Container{container}
	for _, c := range cc {
		if c != nil {
			err = errors.Join(err, c.Terminate(ctx))
		}
	}
	return err
}
