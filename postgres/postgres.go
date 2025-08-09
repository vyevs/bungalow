package postgres

import (
	"context"
	"fmt"

	"github.com/vyevs/bungalow/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

func NewClient(ctx context.Context, domain, port, db, user, password string) (*Client, error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?search_path=bungalow&connect_timeout=5&pool_max_conns=5",
		user, password, domain, port, db)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("new conn pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return &Client{
		pool: pool,
	}, nil
}

func (c Client) InsertPerson(ctx context.Context, person domain.Person) error {
	_, err := c.pool.Exec(ctx, "INSERT INTO person (firstName, lastName) VALUES ($1, $2)", person.FirstName, person.LastName)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (c Client) Close() {
	c.pool.Close()
}

func (c Client) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
