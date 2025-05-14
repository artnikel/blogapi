package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/artnikel/blogapi/internal/config"
	"github.com/artnikel/blogapi/internal/model"
	"github.com/caarlos0/env"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

var pgRepo *PgRepository

func SetupPostgres() (*pgxpool.Pool, func(), error) {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, nil, fmt.Errorf("Failed to parse config: %v", err)
	}
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("could not construct pool: %w", err)
	}
	resource, err := pool.Run("postgres", "latest", []string{
		fmt.Sprintf("POSTGRES_USER=%s", cfg.BlogPostgresUser),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", cfg.BlogPostgresPassword),
		fmt.Sprintf("POSTGRES_DB=%s", cfg.BlogPostgresDB),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s",
		cfg.BlogPostgresUser,
		cfg.BlogPostgresPassword,
		resource.GetPort("5432"),
		cfg.BlogPostgresDB)
	conf, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse dbURL: %w", err)
	}
	dbpool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect pgxpool: %w", err)
	}
	cleanup := func() {
		dbpool.Close()
		pool.Purge(resource)
	}
	return dbpool, cleanup, nil
}

func TestMain(m *testing.M) {
	dbpool, cleanupPgx, err := SetupPostgres()
	if err != nil {
		fmt.Println("Could not construct the pool: ", err)
		cleanupPgx()
		os.Exit(1)
	}
	pgRepo = NewPgRepository(dbpool)
	exitCode := m.Run()
	cleanupPgx()
	os.Exit(exitCode)
}

var testBlog = model.Blog{
	BlogID:  uuid.New(),
	UserID:  uuid.New(),
	Title:   "testtitle",
	Content: "testcontent",
}

func Test_Create(t *testing.T) {
	err := pgRepo.Create(context.Background(), &testBlog)
	require.NoError(t, err)
	blog, err := pgRepo.Get(context.Background(), testBlog.BlogID)
	require.NoError(t, err)
	require.Equal(t, blog.UserID, testBlog.UserID)
	require.Equal(t, blog.Title, testBlog.Title)
	require.Equal(t, blog.Content, testBlog.Content)
}

func Test_CreateDuplicate(t *testing.T) {
	err := pgRepo.Create(context.Background(), &testBlog)
	require.Error(t, err)
}

func Test_CreateContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	time.Sleep(1 * time.Second)
	defer cancel()
	err := pgRepo.Create(ctx, &testBlog)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
}

func Test_GetNotFound(t *testing.T) {
	_, err := pgRepo.Get(context.Background(), uuid.New())
	require.Error(t, err)
}

func Test_GetContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	time.Sleep(1 * time.Second)
	defer cancel()
	_, err := pgRepo.Get(ctx, testBlog.BlogID)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
}

func Test_GetAll(t *testing.T) {
	allBlogs, err := pgRepo.GetAll(context.Background())
	require.NoError(t, err)
	var lenBlogs int
	err = pgRepo.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM blog").Scan(&lenBlogs)
	require.NoError(t, err)
	require.Equal(t, len(allBlogs), lenBlogs)
}

func Test_Update(t *testing.T) {
	testBlog.Title = "testtitle2"
	testBlog.Content = "testcontent2"
	err := pgRepo.Update(context.Background(), &testBlog)
	require.NoError(t, err)
	blog, err := pgRepo.Get(context.Background(), testBlog.BlogID)
	require.NoError(t, err)
	require.Equal(t, blog.UserID, testBlog.UserID)
	require.Equal(t, blog.Title, testBlog.Title)
	require.Equal(t, blog.Content, testBlog.Content)
}

func Test_UpdateContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	time.Sleep(1 * time.Second)
	defer cancel()
	err := pgRepo.Update(ctx, &testBlog)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
}

func Test_Delete(t *testing.T) {
	err := pgRepo.Delete(context.Background(), testBlog.BlogID)
	require.NoError(t, err)
	_, err = pgRepo.Get(context.Background(), testBlog.BlogID)
	require.Error(t, err)
}

func Test_DeleteContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	time.Sleep(1 * time.Second)
	defer cancel()
	err := pgRepo.Delete(ctx, testBlog.BlogID)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
}

func Test_DeleteByUserID(t *testing.T) {
	err := pgRepo.DeleteByUserID(context.Background(), testBlog.UserID)
	require.NoError(t, err)
	_, err = pgRepo.Get(context.Background(), testBlog.BlogID)
	require.Error(t, err)
}

func Test_DeleteNonExistent(t *testing.T) {
	err := pgRepo.Delete(context.Background(), uuid.New())
	require.NoError(t, err, "Expected no error when deleting non-existent record")
}

func Test_GetByUserIDNotFound(t *testing.T) {
	blogs, err := pgRepo.GetByUserID(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Empty(t, blogs)
}

func Test_DeleteByUserIDAllEntries(t *testing.T) {
	newUserID := uuid.New()
	blog1 := model.Blog{BlogID: uuid.New(), UserID: newUserID, Title: "title1", Content: "content1"}
	blog2 := model.Blog{BlogID: uuid.New(), UserID: newUserID, Title: "title2", Content: "content2"}
	_ = pgRepo.Create(context.Background(), &blog1)
	_ = pgRepo.Create(context.Background(), &blog2)

	err := pgRepo.DeleteByUserID(context.Background(), newUserID)
	require.NoError(t, err)

	blogs, err := pgRepo.GetByUserID(context.Background(), newUserID)
	require.NoError(t, err)
	require.Empty(t, blogs)
}

func Test_GetAllEmptyAfterDeletion(t *testing.T) {
	_, err := pgRepo.pool.Exec(context.Background(), "TRUNCATE blog RESTART IDENTITY")
	require.NoError(t, err)
	allBlogs, err := pgRepo.GetAll(context.Background())
	require.NoError(t, err)
	require.Empty(t, allBlogs)
}
