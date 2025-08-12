package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/vyevs/bungalow/clients/postgres"
	"github.com/vyevs/bungalow/domain"
)

type PeopleHandler struct {
	Cli *postgres.Client
}

func (h *PeopleHandler) CreatePerson(rw http.ResponseWriter, req *http.Request) {
	status, id, err := h.createPerson(req)

	rw.WriteHeader(status)
	if status == http.StatusOK {
		var resp struct {
			ID int `json:"id"`
		}
		resp.ID = id
		_ = json.NewEncoder(rw).Encode(resp)
	} else {
		encodeErr(rw, err)
	}
}

type errResp struct {
	Message string `json:"message"`
}

// returns (http status code, new person's id, error)
func (h *PeopleHandler) createPerson(req *http.Request) (int, int, error) {
	var p domain.Person
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
		return http.StatusBadRequest, 0, fmt.Errorf("failed to decode request body: %w", err)
	}

	if p.FirstName == "" {
		return http.StatusBadRequest, 0, errors.New("missing first name")
	}
	if p.LastName == "" {
		return http.StatusBadRequest, 0, errors.New("missing last name")
	}

	ctx := req.Context()

	id, err := h.Cli.CreatePerson(ctx, p)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("failed to create person: %w", err)
	}

	return http.StatusOK, id, nil
}

func (h *PeopleHandler) GetPerson(rw http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	status, person, err := h.getPerson(req.Context(), idStr)

	rw.WriteHeader(status)
	if status == http.StatusOK {
		_ = json.NewEncoder(rw).Encode(person)

	} else {
		encodeErr(rw, err)
	}
}

func (h *PeopleHandler) getPerson(ctx context.Context, idStr string) (int, domain.Person, error) {
	id, err := strconv.Atoi(idStr)
	// If provided a non-integer ID, we return Not Found since all our IDs are integers.
	if err != nil {
		return http.StatusNotFound, domain.Person{}, errors.New("not found")
	}

	p, err := h.Cli.GetPerson(ctx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return http.StatusNotFound, domain.Person{}, fmt.Errorf("not found")
		}
		return http.StatusInternalServerError, domain.Person{}, fmt.Errorf("failed to get person: %w", err)
	}

	return http.StatusOK, p, nil
}

func (h *PeopleHandler) DeletePerson(rw http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	status, err := h.deletePerson(req.Context(), idStr)

	rw.WriteHeader(http.StatusOK)
	if status != http.StatusOK {
		encodeErr(rw, err)
	}
}

// returns (http status code, error)
func (h *PeopleHandler) deletePerson(ctx context.Context, idStr string) (int, error) {
	// If provided a non-integer ID, we return Not Found since all our IDs are integers.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return http.StatusNotFound, errors.New("not found")
	}

	if err := h.Cli.DeletePerson(ctx, id); err != nil {
		if err == postgres.ErrNotFound {
			return http.StatusNotFound, errors.New("not found")
		}
		return http.StatusInternalServerError, fmt.Errorf("failed to delete person: %w", err)
	}

	return http.StatusOK, nil
}

func encodeErr(w io.Writer, err error) {
	errResp := errResp{
		Message: err.Error(),
	}
	_ = json.NewEncoder(w).Encode(errResp)
}
