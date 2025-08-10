package postgres

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
)

const defaultVersionTable string = "schema_version_non_default"

var defaultMigrator = &migrate.MigratorOptions{
	DisableTx: false,
}

type FS struct {
	URL string
	FS  fs.FS
}

// Up runs the up migrations on the database.
func (fsys FS) Up(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, fsys.URL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		ctx = context.WithoutCancel(ctx)
		conn.Close(ctx)
	}()
	migrator, err := migrate.NewMigratorEx(ctx, conn, defaultVersionTable, defaultMigrator)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	err = migrator.LoadMigrations(fsys.FS)
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}
	err = migrator.Migrate(ctx) // up
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	return nil
}

// Down runs the down migrations on the database.
func (fsys FS) Down(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, fsys.URL)
	if err != nil {
		return err
	}
	defer func() {
		ctx = context.WithoutCancel(ctx)
		conn.Close(ctx)
	}()
	migrator, err := migrate.NewMigratorEx(ctx, conn, defaultVersionTable, defaultMigrator)
	if err != nil {
		return err
	}
	err = migrator.LoadMigrations(fsys.FS)
	if err != nil {
		return err
	}
	err = migrator.MigrateTo(ctx, 0) // down
	if err != nil {
		return err
	}
	return nil
}

// Version runs the migrations up to a version.
func (fsys FS) Version(ctx context.Context, ver int32) error {
	conn, err := pgx.Connect(ctx, fsys.URL)
	if err != nil {
		return err
	}
	defer func() {
		ctx = context.WithoutCancel(ctx)
		conn.Close(ctx)
	}()
	migrator, err := migrate.NewMigratorEx(ctx, conn, defaultVersionTable, defaultMigrator)
	if err != nil {
		return err
	}
	// needed?
	err = migrator.LoadMigrations(fsys.FS)
	if err != nil {
		return err
	}
	err = migrator.MigrateTo(ctx, ver) // down
	if err != nil {
		return err
	}
	return nil
}
