package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Health struct {
	Postgres Pinger // postgres
}

type response struct {
	status   int
	Postgres string `json:"postgres"`
}

func (h Health) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	var resp response
	if err := h.Postgres.Ping(ctx); err != nil {
		resp.status = http.StatusInternalServerError
		resp.Postgres = err.Error()
	} else {
		resp.status = http.StatusOK
		resp.Postgres = "healthy"
	}

	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(resp.status)
	_ = json.NewEncoder(rw).Encode(resp)
}
