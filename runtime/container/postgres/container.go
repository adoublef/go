package postgres

import (
	"cmp"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const DefaultImage = "postgres:17-alpine"

type (
	Pool = pgxpool.Pool
)

func Run(ctx context.Context, image string) (*Container, error) {
	c, err := postgres.Run(ctx,
		cmp.Or(image, DefaultImage),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres instance: %w", err)
	}
	return &Container{c}, nil
}

type Container struct{ *postgres.PostgresContainer }

// ConnectionPool creates a new [pgxpool.Pool].
func (c *Container) ConnectionPool(ctx context.Context, args ...string) (*pgxpool.Pool, error) {
	url, err := c.ConnectionString(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to return connection string: %w", err)
	}
	// todo: proxy
	rwc, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to return postgres pool: %w", err)
	}
	return rwc, nil
}

// ConnectionConfig creates a new [pgxpool.Config].
func (c *Container) ConnectionConfig(ctx context.Context, args ...string) (*pgxpool.Config, func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error), error) {
	url, err := c.ConnectionString(ctx, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to return connection string: %w", err)
	}
	// todo: proxy
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to return postgres config: %w", err)
	}
	return cfg, pgxpool.NewWithConfig, nil
}
