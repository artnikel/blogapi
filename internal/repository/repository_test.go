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

var testUser = model.User{
	ID:       uuid.New(),
	Username: "testuserrefresh",
	Password: []byte("password"),
	Admin:    false,
}

func Test_CreateBlog(t *testing.T) {
	ctx := context.Background()
	testBlog.BlogID = uuid.New()
	err := pgRepo.Create(ctx, &testBlog)
	require.NoError(t, err)

	fetchedBlog, err := pgRepo.Get(ctx, testBlog.BlogID)
	require.NoError(t, err)
	require.Equal(t, testBlog.Title, fetchedBlog.Title)
	require.Equal(t, testBlog.Content, fetchedBlog.Content)
}

func Test_CreateBlog_Duplicate(t *testing.T) {
	ctx := context.Background()
	testBlog.BlogID = uuid.New()
	err := pgRepo.Create(ctx, &testBlog)
	require.NoError(t, err)

	err = pgRepo.Create(ctx, &testBlog)
	require.Error(t, err)
}

func Test_CreateBlog_ContextTimeout(t *testing.T) {
	testBlog.BlogID = uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	time.Sleep(1 * time.Second)
	defer cancel()
	err := pgRepo.Create(ctx, &testBlog)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
}

func Test_GetBlog_NotFound(t *testing.T) {
	_, err := pgRepo.Get(context.Background(), uuid.New())
	require.Error(t, err)
}

func Test_GetAllBlogs(t *testing.T) {
	ctx := context.Background()
	firstblogs, err := pgRepo.GetAll(ctx)
	require.NoError(t, err)

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

	_ = pgRepo.Create(ctx, &testBlog1)
	_ = pgRepo.Create(ctx, &testBlog2)

	blogs, err := pgRepo.GetAll(ctx)
	require.NoError(t, err)
	require.Equal(t, len(blogs), len(firstblogs)+2)
}

func Test_UpdateBlog(t *testing.T) {
	ctx := context.Background()
	testBlog.BlogID = uuid.New()
	_ = pgRepo.Create(ctx, &testBlog)

	testBlog.Title = "Updated Title"
	testBlog.Content = "Updated Content"
	err := pgRepo.Update(ctx, &testBlog)
	require.NoError(t, err)

	updatedBlog, err := pgRepo.Get(ctx, testBlog.BlogID)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", updatedBlog.Title)
	require.Equal(t, "Updated Content", updatedBlog.Content)
}

func Test_DeleteBlog(t *testing.T) {
	ctx := context.Background()
	testBlog.BlogID = uuid.New()

	_ = pgRepo.Create(ctx, &testBlog)

	err := pgRepo.Delete(ctx, testBlog.BlogID)
	require.NoError(t, err)

	_, err = pgRepo.Get(ctx, testBlog.BlogID)
	require.Error(t, err)
}

func Test_DeleteBlogsByUserID(t *testing.T) {
	ctx := context.Background()
	testBlog.BlogID = uuid.New()

	_ = pgRepo.Create(ctx, &testBlog)

	err := pgRepo.DeleteBlogsByUserID(ctx, testBlog.UserID)
	require.NoError(t, err)

	_, err = pgRepo.Get(ctx, testBlog.BlogID)
	require.Error(t, err)
}

func Test_GetByUserID_NoBlogs(t *testing.T) {
	blogs, err := pgRepo.GetByUserID(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Empty(t, blogs)
}

func Test_SignUp(t *testing.T) {
	ctx := context.Background()
	testUser.Username = "testusername"
	testUser.ID = uuid.New()

	err := pgRepo.SignUp(ctx, &testUser)
	require.NoError(t, err)

	id, password, admin, err := pgRepo.GetDataByUsername(ctx, testUser.Username)
	require.NoError(t, err)
	require.Equal(t, testUser.ID, id)
	require.Equal(t, testUser.Password, password)
	require.Equal(t, testUser.Admin, admin)
}

func Test_SignUp_ExistingUser(t *testing.T) {
	ctx := context.Background()
	testUser.Username = "testusername2"
	testUser.ID = uuid.New()

	err := pgRepo.SignUp(ctx, &testUser)
	require.NoError(t, err)

	err = pgRepo.SignUp(ctx, &testUser)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrExist)
}

func Test_SignUp_NilUser(t *testing.T) {
	err := pgRepo.SignUp(context.Background(), nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNil)
}

func Test_GetDataByUsername_NotFound(t *testing.T) {
	_, _, _, err := pgRepo.GetDataByUsername(context.Background(), "nonexistent")
	require.Error(t, err)
}

func Test_GetRefreshTokenByID(t *testing.T) {
	ctx := context.Background()
	testUser.Username = "testusername3"
	testUser.ID = uuid.New()

	_ = pgRepo.SignUp(ctx, &testUser)
	testUser.RefreshToken = "test_refresh_token"
	_ = pgRepo.AddRefreshToken(ctx, &testUser)

	storedToken, err := pgRepo.GetRefreshTokenByID(ctx, testUser.ID)
	require.NoError(t, err)
	require.Equal(t, "test_refresh_token", storedToken)
}

func Test_GetRefreshTokenByID_NotFound(t *testing.T) {
	_, err := pgRepo.GetRefreshTokenByID(context.Background(), uuid.New())
	require.Error(t, err)
}

func Test_AddRefreshToken(t *testing.T) {
	ctx := context.Background()
	testUser.Username = "testusername4"

	_ = pgRepo.SignUp(ctx, &testUser)

	newToken := "new_refresh_token"
	testUser.RefreshToken = newToken
	err := pgRepo.AddRefreshToken(ctx, &testUser)
	require.NoError(t, err)

	storedToken, err := pgRepo.GetRefreshTokenByID(ctx, testUser.ID)
	require.NoError(t, err)
	require.Equal(t, newToken, storedToken)
}

func Test_DeleteUserByID(t *testing.T) {
	ctx := context.Background()

	testUser.Username = "testusername5"
	testUser.ID = uuid.New()

	err := pgRepo.SignUp(ctx, &testUser)
	require.NoError(t, err)

	err = pgRepo.DeleteUserByID(ctx, testUser.ID)
	require.NoError(t, err)

	_, _, _, err = pgRepo.GetDataByUsername(ctx, testUser.Username)
	require.Error(t, err)
}

func Test_DeleteUserByID_AdminUser(t *testing.T) {
	ctx := context.Background()
	testUser.Username = "testusername6"
	testUser.ID = uuid.New()
	testUser.Admin = true

	err := pgRepo.SignUp(ctx, &testUser)
	require.NoError(t, err)

	err = pgRepo.DeleteUserByID(ctx, testUser.ID)
	require.Error(t, err)

	id, _, _, err := pgRepo.GetDataByUsername(ctx, testUser.Username)
	require.NoError(t, err)
	require.Equal(t, testUser.ID, id)
}

func Test_DeleteUserByID_UserNotFound(t *testing.T) {
	err := pgRepo.DeleteUserByID(context.Background(), uuid.New())
	require.Error(t, err)
}
