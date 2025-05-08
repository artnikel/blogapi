package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/artnikel/blogapi/internal/config"
	// "github.com/artnikel/blogapi/internal/handler"
	// "github.com/artnikel/blogapi/internal/repository"
	// "github.com/artnikel/blogapi/internal/service"
	"github.com/caarlos0/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	//"gopkg.in/go-playground/validator.v9"
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
	//v := validator.New()
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	pool, err := connectPostgres()
	if err != nil {
		fmt.Printf("Failed to connect to Postgres: %v", err)
	}
	defer pool.Close()

	// repoPostgres := repository.NewPgRepository(pool)
	// blogService := service.NewBlogService(repoPostgres)
	// userService := service.NewUserService(repoPostgres, &cfg)
	// blogHandler := handler.NewEntityBlog(blogService, v)
	// userHandler := handler.NewEntityUser(userService, v)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
