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

func clearDB(t *testing.T) {
	_, err := pgRepo.pool.Exec(context.Background(), "TRUNCATE users,blog RESTART IDENTITY")
	require.NoError(t, err)
}

func Test_CreateBlog(t *testing.T) {
	clearDB(t)

	testBlog := model.Blog{
		BlogID:  uuid.New(),
		UserID:  uuid.New(),
		Title:   "Test Blog",
		Content: "This is a test blog",
	}

	err := pgRepo.Create(context.Background(), &testBlog)
	require.NoError(t, err)

	fetchedBlog, err := pgRepo.Get(context.Background(), testBlog.BlogID)
	require.NoError(t, err)
	require.Equal(t, testBlog.Title, fetchedBlog.Title)
	require.Equal(t, testBlog.Content, fetchedBlog.Content)
}

func Test_CreateBlog_Duplicate(t *testing.T) {
	clearDB(t)

	testBlog := model.Blog{
		BlogID:  uuid.New(),
		UserID:  uuid.New(),
		Title:   "Test Blog",
		Content: "This is a test blog",
	}

	err := pgRepo.Create(context.Background(), &testBlog)
	require.NoError(t, err)

	err = pgRepo.Create(context.Background(), &testBlog)
	require.Error(t, err)
}

func Test_CreateBlog_ContextTimeout(t *testing.T) {
	clearDB(t)

	testBlog := model.Blog{
		BlogID:  uuid.New(),
		UserID:  uuid.New(),
		Title:   "Test Blog Timeout",
		Content: "This is a test blog for timeout",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	time.Sleep(1 * time.Second)
	defer cancel()
	err := pgRepo.Create(ctx, &testBlog)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
}

func Test_GetBlog_NotFound(t *testing.T) {
	clearDB(t)

	_, err := pgRepo.Get(context.Background(), uuid.New())
	require.Error(t, err)
}

func Test_GetAllBlogs(t *testing.T) {
	clearDB(t)

	testBlog1 := model.Blog{
		BlogID:  uuid.New(),
		UserID:  uuid.New(),
		Title:   "First Blog",
		Content: "Content of first blog",
	}
	testBlog2 := model.Blog{
		BlogID:  uuid.New(),
		UserID:  uuid.New(),
		Title:   "Second Blog",
		Content: "Content of second blog",
	}

	_ = pgRepo.Create(context.Background(), &testBlog1)
	_ = pgRepo.Create(context.Background(), &testBlog2)

	blogs, err := pgRepo.GetAll(context.Background())
	require.NoError(t, err)
	require.Len(t, blogs, 2)
}

func Test_UpdateBlog(t *testing.T) {
	clearDB(t)

	testBlog := model.Blog{
		BlogID:  uuid.New(),
		UserID:  uuid.New(),
		Title:   "Original Title",
		Content: "Original Content",
	}

	_ = pgRepo.Create(context.Background(), &testBlog)

	testBlog.Title = "Updated Title"
	testBlog.Content = "Updated Content"
	err := pgRepo.Update(context.Background(), &testBlog)
	require.NoError(t, err)

	updatedBlog, err := pgRepo.Get(context.Background(), testBlog.BlogID)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", updatedBlog.Title)
	require.Equal(t, "Updated Content", updatedBlog.Content)
}

func Test_DeleteBlog(t *testing.T) {
	clearDB(t)

	testBlog := model.Blog{
		BlogID:  uuid.New(),
		UserID:  uuid.New(),
		Title:   "To be deleted",
		Content: "This blog will be deleted",
	}

	_ = pgRepo.Create(context.Background(), &testBlog)

	err := pgRepo.Delete(context.Background(), testBlog.BlogID)
	require.NoError(t, err)

	_, err = pgRepo.Get(context.Background(), testBlog.BlogID)
	require.Error(t, err)
}

func Test_GetByUserID_NoBlogs(t *testing.T) {
	clearDB(t)

	blogs, err := pgRepo.GetByUserID(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Empty(t, blogs)
}

func Test_SignUp(t *testing.T) {
	clearDB(t)

	testUser := model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: []byte("password"),
		Admin:    false,
	}

	err := pgRepo.SignUp(context.Background(), &testUser)
	require.NoError(t, err)

	id, password, admin, err := pgRepo.GetDataByUsername(context.Background(), testUser.Username)
	require.NoError(t, err)
	require.Equal(t, testUser.ID, id)
	require.Equal(t, testUser.Password, password)
	require.Equal(t, testUser.Admin, admin)
}

func Test_SignUp_ExistingUser(t *testing.T) {
	clearDB(t)

	testUser := model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: []byte("password"),
		Admin:    false,
	}

	err := pgRepo.SignUp(context.Background(), &testUser)
	require.NoError(t, err)

	err = pgRepo.SignUp(context.Background(), &testUser)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrExist)
}

func Test_SignUp_NilUser(t *testing.T) {
	clearDB(t)

	err := pgRepo.SignUp(context.Background(), nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNil)
}

func Test_GetDataByUsername_NotFound(t *testing.T) {
	clearDB(t)

	_, _, _, err := pgRepo.GetDataByUsername(context.Background(), "nonexistent")
	require.Error(t, err)
}

func Test_GetRefreshTokenByID(t *testing.T) {
	clearDB(t)

	testUser := model.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Password:     []byte("password"),
		Admin:        false,
		RefreshToken: "test_refresh_token",
	}

	_ = pgRepo.SignUp(context.Background(), &testUser)
	_ = pgRepo.AddRefreshToken(context.Background(), &testUser)

	storedToken, err := pgRepo.GetRefreshTokenByID(context.Background(), testUser.ID)
	require.NoError(t, err)
	require.Equal(t, "test_refresh_token", storedToken)
}

func Test_GetRefreshTokenByID_NotFound(t *testing.T) {
	clearDB(t)

	_, err := pgRepo.GetRefreshTokenByID(context.Background(), uuid.New())
	require.Error(t, err)
}

func Test_AddRefreshToken(t *testing.T) {
	clearDB(t)

	testUser := model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: []byte("password"),
		Admin:    false,
	}

	_ = pgRepo.SignUp(context.Background(), &testUser)

	newToken := "new_refresh_token"
	testUser.RefreshToken = newToken
	err := pgRepo.AddRefreshToken(context.Background(), &testUser)
	require.NoError(t, err)

	storedToken, err := pgRepo.GetRefreshTokenByID(context.Background(), testUser.ID)
	require.NoError(t, err)
	require.Equal(t, newToken, storedToken)
}
