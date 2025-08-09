package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vyevs/bungalow/handlers"
	"github.com/vyevs/bungalow/handlers/middleware"
	"github.com/vyevs/bungalow/postgres"
)

func main() {
	debug.SetMemoryLimit(4 * (1 << 30)) // 4 GiB
	debug.SetGCPercent(-1)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logr := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logr)

	err := mainWithCtx(ctx, cancel)
	if err != nil {
		slog.Error("something went wrong", slog.Any("error", err))
	}

	slog.Info("shut down gracefully")
}

func mainWithCtx(ctx context.Context, cancel context.CancelFunc) error {
	pgCli, err := newPostgresClient()
	if err != nil {
		return fmt.Errorf("postgres new client: %w", err)
	}
	defer func() {
		pgCli.Close()
		slog.Info("postgres client successfully closed")
	}()

	initMetrics()

	mws := middleware.Middlewares{
		middleware.Recover,
		middleware.Metrics,
		middleware.LogRequest,
	}

	http.Handle("/metrics", mws.Wrap(promhttp.Handler()))

	healthHandler := handlers.Health{
		Postgres: pgCli,
	}
	http.Handle("/health", mws.Wrap(healthHandler))

	server := http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	var wg sync.WaitGroup

	done := func() {
		wg.Done()
		cancel()
	}

	wg.Add(1)
	go func() {
		defer done()

		log.Print("started HTTP server")
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Print("HTTP server shut down gracefully")
		} else {
			log.Printf("HTTP server shut down due to error: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer done()

		for {
			select {
			case <-time.Tick(1 * time.Second):
				log.Print("HEY THERE SAILOR!!")
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()

	{
		serverShutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(serverShutdownCtx); err != nil {
			slog.Error("server.Shutdown error", slog.Any("error", err))
		}
	}

	return nil
}

func newPostgresClient() (*postgres.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return postgres.NewClient(ctx, "localhost", "5432", "postgres", "postgres", "password")
}

func initMetrics() {
	middleware.InitMetrics()
}
