// A main package
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
	"syscall"

	"github.com/artnikel/blogapi/internal/config"
	"github.com/artnikel/blogapi/internal/constants"
	"github.com/artnikel/blogapi/internal/handler"
	customMiddleware "github.com/artnikel/blogapi/internal/middleware"
	"github.com/artnikel/blogapi/internal/repository"
	"github.com/artnikel/blogapi/internal/service"
	"github.com/caarlos0/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/go-playground/validator.v9"
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
	v := validator.New()

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	pool, err := connectPostgres()
	if err != nil {
		fmt.Printf("Failed to connect to Postgres: %v", err)
	}
	defer pool.Close()

	repoPostgres := repository.NewPgRepository(pool)
	blogService := service.NewBlogService(repoPostgres)
	userService := service.NewUserService(repoPostgres, &cfg)
	handlers := handler.NewHandler(blogService, userService, v)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/blog", handlers.Create, customMiddleware.JWTMiddleware(&cfg))
	e.GET("/blog/:id", handlers.Get, customMiddleware.JWTMiddleware(&cfg))
	e.DELETE("/blog/:id", handlers.Delete, customMiddleware.JWTMiddleware(&cfg))
	e.DELETE("/blogs/user/:id", handlers.DeleteBlogsByUserID, customMiddleware.JWTMiddleware(&cfg))
	e.PUT("/blog", handlers.Update, customMiddleware.JWTMiddleware(&cfg))
	e.GET("/blogs", handlers.GetAll, customMiddleware.JWTMiddleware(&cfg))
	e.GET("/blogs/user/:id", handlers.GetByUserID, customMiddleware.JWTMiddleware(&cfg))

	e.POST("/signup", handlers.SignUpUser)
	e.POST("/signupadmin", handlers.SignUpAdmin, customMiddleware.JWTMiddleware(&cfg))
	e.POST("/login", handlers.Login)
	e.POST("/refresh", handlers.Refresh)
	e.DELETE("/user/:id", handlers.DeleteUserByID, customMiddleware.JWTMiddleware(&cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(":" + cfg.BlogServerPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start server", "error", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.ServerTimeout)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown error %v", err)
	}
	log.Println("Server gracefully stopped")
}
