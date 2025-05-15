package service

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/artnikel/blogapi/internal/config"
	"github.com/artnikel/blogapi/internal/model"
	"github.com/artnikel/blogapi/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_SignUp(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	cfg := &config.Config{BlogTokenSignature: "secret"}
	svc := NewUserService(mockRepo, cfg)

	user := &model.User{
		Username: "testuser",
		Password: []byte("password123"),
	}

	mockRepo.EXPECT().
		SignUp(mock.Anything, mock.AnythingOfType("*model.User")).
		Return(nil).
		Run(func(ctx context.Context, u *model.User) {
			require.NotEqual(t, []byte("password123"), u.Password)
		})

	err := svc.SignUp(context.Background(), user)
	require.NoError(t, err)
}

func TestUserService_Login(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	cfg := &config.Config{BlogTokenSignature: "secret"}
	svc := NewUserService(mockRepo, cfg)

	userID := uuid.New()
	password := []byte("password123")
	hashedPass, _ := svc.HashPassword(password)

	user := &model.User{
		Username: "testuser",
		Password: password,
	}

	mockRepo.EXPECT().
		GetDataByUsername(mock.Anything, user.Username).
		Return(userID, hashedPass, true, nil)

	mockRepo.EXPECT().
		AddRefreshToken(mock.Anything, mock.AnythingOfType("*model.User")).
		Return(nil).
		Run(func(ctx context.Context, u *model.User) {
			require.NotEmpty(t, u.RefreshToken)
		})

	tokens, err := svc.Login(context.Background(), user)
	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	require.Equal(t, userID, user.ID)
	require.True(t, user.Admin)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	cfg := &config.Config{BlogTokenSignature: "secret"}
	svc := NewUserService(mockRepo, cfg)

	userID := uuid.New()
	password := []byte("correct_password")
	hashedPass, _ := svc.HashPassword(password)

	user := &model.User{
		Username: "testuser",
		Password: []byte("wrong_password"),
	}

	mockRepo.EXPECT().
		GetDataByUsername(mock.Anything, user.Username).
		Return(userID, hashedPass, false, nil)

	tokens, err := svc.Login(context.Background(), user)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CheckPasswordHash")
	require.Empty(t, tokens.AccessToken)
}

func TestUserService_Refresh(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	cfg := &config.Config{BlogTokenSignature: "secret"}
	svc := NewUserService(mockRepo, cfg)

	userID := uuid.New()
	isAdmin := true

	tokenPair, err := svc.GenerateTokenPair(userID, isAdmin)
	require.NoError(t, err)

	sum := sha256.Sum256([]byte(tokenPair.RefreshToken))
	hashedRefreshToken, err := svc.HashPassword(sum[:]) 
	require.NoError(t, err)

	mockRepo.EXPECT().
		GetRefreshTokenByID(mock.Anything, userID).
		Return(string(hashedRefreshToken), nil)

	mockRepo.EXPECT().
		AddRefreshToken(mock.Anything, mock.AnythingOfType("*model.User")).
		Return(nil).
		Run(func(ctx context.Context, u *model.User) {
			require.NotEmpty(t, u.RefreshToken)
		})

	newTokenPair, err := svc.Refresh(context.Background(), tokenPair)
	require.NoError(t, err)
	require.NotEmpty(t, newTokenPair.AccessToken)
	require.NotEmpty(t, newTokenPair.RefreshToken)
}

func TestUserService_Refresh_InvalidToken(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository(t)
	cfg := &config.Config{BlogTokenSignature: "secret"}
	svc := NewUserService(mockRepo, cfg)

	userID := uuid.New()
	isAdmin := true

	tokenPair, err := svc.GenerateTokenPair(userID, isAdmin)
	require.NoError(t, err)

	mockRepo.EXPECT().
		GetRefreshTokenByID(mock.Anything, userID).
		Return("some_invalid_hash", nil)

	_, err = svc.Refresh(context.Background(), tokenPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "CheckPasswordHash error")
}

