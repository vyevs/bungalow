package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/vyevs/bungalow/domain"
	"github.com/vyevs/bungalow/postgres"
)

func main() {
	if err := createPeople(); err != nil {
		log.Printf("error: %v", err)
	}
}

func createPeople() error {
	pgCli, err := newPostgresClient()
	if err != nil {
		return fmt.Errorf("creating postgres client: %w", err)
	}
	defer pgCli.Close()

	for range 10 {
		person := domain.Person{
			FirstName: gofakeit.FirstName(),
			LastName:  gofakeit.LastName(),
		}
		err := pgCli.InsertPerson(context.Background(), person)
		if err != nil {
			return fmt.Errorf("failed to insert person: %w", err)
		}
	}

	return nil
}

func newPostgresClient() (*postgres.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return postgres.NewClient(ctx, "localhost", "5432", "postgres", "postgres", "password")
}
