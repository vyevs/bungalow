// Package postgres implements functionality to interface with postgres using the bungalow domain model.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/vyevs/bungalow/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("not found")
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

func (c Client) CreatePerson(ctx context.Context, person domain.Person) (int, error) {
	row := c.pool.QueryRow(ctx, "INSERT INTO person (firstName, lastName) VALUES ($1, $2) RETURNING id", person.FirstName, person.LastName)

	var id int
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}

	return id, nil
}

func (c Client) GetPerson(ctx context.Context, id int) (domain.Person, error) {
	row := c.pool.QueryRow(ctx, "SELECT firstName, lastName FROM person WHERE id = $1", id)

	var p domain.Person
	if err := row.Scan(&p.FirstName, &p.LastName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Person{}, ErrNotFound
		}
		return domain.Person{}, fmt.Errorf("scan: %w", err)
	}
	p.ID = id

	return p, nil
}

func (c Client) DeletePerson(ctx context.Context, id int) error {
	tag, err := c.pool.Exec(ctx, "DELETE FROM person WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Close closes all the client's connections to postgres.
// The client should not be used after a call to Close.
func (c Client) Close() {
	c.pool.Close()
}

// Ping pings the postgres server to verify connectivity.
func (c Client) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
