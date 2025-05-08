package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/artnikel/blogapi/internal/config"
	"github.com/caarlos0/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func connectPostgres() (*pgxpool.Pool, error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	conf, err := pgxpool.ParseConfig(cfg.BlogPostgresPath)
	if err != nil {
		return nil, fmt.Errorf("error in method pgxpool.ParseConfig: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("error in method pgxpool.NewWithConfig: %v", err)
	}
	return pool, nil
}

func main() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	pool, err := connectPostgres()
	if err != nil {
		fmt.Printf("Failed to connect to Postgres: %v", err)
	}
	defer pool.Close()

	//repoPostgres := repository.NewPgRepository(pool)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
