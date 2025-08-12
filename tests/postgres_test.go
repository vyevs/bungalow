package tests

import (
	"context"
	"math"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"

	"github.com/vyevs/bungalow/clients/postgres"
	"github.com/vyevs/bungalow/domain"

	. "github.com/onsi/gomega"
)

func TestPostgres(t *testing.T) {
	RegisterTestingT(t)

	cli, err := newPostgresClient()
	Expect(err).ToNot(HaveOccurred())

	firstName := strconv.Itoa(rand.N(math.MaxInt))
	lastName := strconv.Itoa(rand.N(math.MaxInt))

	id, err := cli.CreatePerson(context.Background(), domain.Person{
		FirstName: firstName,
		LastName:  lastName,
	})
	Expect(err).ToNot(HaveOccurred())

	// Verify the person we just created exists.
	p, err := cli.GetPerson(context.Background(), id)
	Expect(err).ToNot(HaveOccurred())
	Expect(p.FirstName).To(Equal(firstName))
	Expect(p.LastName).To(Equal(lastName))

	err = cli.DeletePerson(context.Background(), id)
	Expect(err).ToNot(HaveOccurred())

	// Verify the person we just deleted does not exist.
	p, err = cli.GetPerson(context.Background(), id)
	Expect(err).To(MatchError(postgres.ErrNotFound))
}

func newPostgresClient() (*postgres.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return postgres.NewClient(ctx, "localhost", "5432", "postgres", "postgres", "password")
}
