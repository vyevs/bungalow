package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/vyevs/bungalow/domain"
	"github.com/vyevs/bungalow/postgres"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := createPeople(ctx); err != nil {
		log.Printf("error: %v", err)
	}
}

func createPeople(ctx context.Context) error {
	pgCli, err := newPostgresClient()
	if err != nil {
		return fmt.Errorf("creating postgres client: %w", err)
	}
	defer pgCli.Close()

	for range 1_000_000 {
		if err := ctx.Err(); err != nil {
			break
		}

		person := domain.Person{
			FirstName: gofakeit.FirstName(),
			LastName:  gofakeit.LastName(),
		}
		err := pgCli.InsertPerson(context.Background(), person)
		if err != nil {
			return fmt.Errorf("failed to insert person: %w", err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func newPostgresClient() (*postgres.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return postgres.NewClient(ctx, "localhost", "5432", "postgres", "postgres", "password")
}
