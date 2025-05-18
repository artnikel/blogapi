// Package service provides the business logic for the auth
package service

import (
	"context"
	"crypto/sha256"
	"time"

	"fmt"

	"github.com/artnikel/blogapi/internal/config"
	"github.com/artnikel/blogapi/internal/constants"
	"github.com/artnikel/blogapi/internal/middleware"
	"github.com/artnikel/blogapi/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository is an interface that contains auth methods
type UserRepository interface {
	SignUp(ctx context.Context, user *model.User) error
	GetDataByUsername(ctx context.Context, username string) (uuid.UUID, []byte, bool, error)
	AddRefreshToken(ctx context.Context, user *model.User) error
	GetRefreshTokenByID(ctx context.Context, id uuid.UUID) (string, error)
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
}

// UserService contains UserRepository interface
type UserService struct {
	rpsUser UserRepository
	cfg     *config.Config
}

// NewUserService accepts UserRepository object and returnes an object of type *UserService
func NewUserService(rpsUser UserRepository, cfg *config.Config) *UserService {
	return &UserService{rpsUser: rpsUser, cfg: cfg}
}

// TokenPair contains an Access and a Refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// SignUp is a method of UserService that calls  method of Repository
func (s *UserService) SignUp(ctx context.Context, user *model.User) error {
	var err error
	user.Password, err = s.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("HashPassword - %w", err)
	}
	err = s.rpsUser.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("rpsUser.SignUp - %w", err)
	}
	return nil
}

// Login is a method of UserService that calls method of Repository
func (s *UserService) Login(ctx context.Context, user *model.User) (*TokenPair, error) {
	id, hash, admin, err := s.rpsUser.GetDataByUsername(ctx, user.Username)
	user.ID = id
	user.Admin = admin
	if err != nil {
		return &TokenPair{}, fmt.Errorf("rpsUser.GetDataByUsername - %w", err)
	}
	verified, err := s.CheckPasswordHash(hash, user.Password)
	if err != nil || !verified {
		return &TokenPair{}, fmt.Errorf("CheckPasswordHash - %w", err)
	}
	tokenPair, err := s.GenerateTokenPair(user.ID, user.Admin)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("GenerateTokenPair - %w", err)
	}
	sum := sha256.Sum256([]byte(tokenPair.RefreshToken))
	hashedRefreshToken, err := s.HashPassword(sum[:])
	if err != nil {
		return &TokenPair{}, fmt.Errorf("HashPassword - %w", err)
	}
	user.RefreshToken = string(hashedRefreshToken)
	err = s.rpsUser.AddRefreshToken(context.Background(), user)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("rpsUser.AddRefreshToken - %w", err)
	}
	return &tokenPair, nil
}

// Refresh is a method of ServiceUser that refreshes access and refresh tokens
func (s *UserService) Refresh(ctx context.Context, tokenPair TokenPair) (TokenPair, error) {
	id, isAdmin, err := s.TokensIDCompare(tokenPair)
	if err != nil {
		return TokenPair{}, fmt.Errorf("TokensIDCompare - %w", err)
	}
	hash, err := s.rpsUser.GetRefreshTokenByID(ctx, id)
	if err != nil {
		return TokenPair{}, fmt.Errorf("rpsUser.GetRefreshTokenByID - %w", err)
	}
	sum := sha256.Sum256([]byte(tokenPair.RefreshToken))
	verified, err := s.CheckPasswordHash([]byte(hash), sum[:])
	if err != nil || !verified {
		return TokenPair{}, fmt.Errorf("CheckPasswordHash error: refreshToken invalid")
	}
	tokenPair, err = s.GenerateTokenPair(id, isAdmin)
	if err != nil {
		return TokenPair{}, fmt.Errorf("GenerateTokenPair - %w", err)
	}
	sum = sha256.Sum256([]byte(tokenPair.RefreshToken))
	hashedRefreshToken, err := s.HashPassword(sum[:])
	if err != nil {
		return TokenPair{}, fmt.Errorf("HashPassword - %w", err)
	}
	var user model.User
	user.RefreshToken = string(hashedRefreshToken)
	user.ID = id
	err = s.rpsUser.AddRefreshToken(context.Background(), &user)
	if err != nil {
		return TokenPair{}, fmt.Errorf("rpsUser.AddRefreshToken - %w", err)
	}
	return tokenPair, nil
}

// DeleteUserByID is a method of UserService that calls  method of Repository
func (s *UserService) DeleteUserByID(ctx context.Context, id uuid.UUID) error {
	err := s.rpsUser.DeleteUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("rpsUser.DeleteUserByID - %w", err)
	}
	return nil
}

// TokensIDCompare compares IDs from refresh and access token for being equal
func (s *UserService) TokensIDCompare(tokenPair TokenPair) (uuid.UUID, bool, error) {
	accessToken, err := middleware.ValidateToken(tokenPair.AccessToken, s.cfg.BlogTokenSignature)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("middleware.validateToken - %w", err)
	}
	var accessID uuid.UUID
	var uuidID uuid.UUID
	var isAdmin bool
	if claims, ok := accessToken.Claims.(jwt.MapClaims); ok && accessToken.Valid {
		uuidID, err = uuid.Parse(claims["id"].(string))
		if err != nil {
			return uuid.Nil, false, fmt.Errorf("uuid.Parse - %w", err)
		}
		isAdmin = claims["isAdmin"].(bool)
		accessID = uuidID
	}
	refreshToken, err := middleware.ValidateToken(tokenPair.RefreshToken, s.cfg.BlogTokenSignature)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("middleware.validateToken - %w", err)
	}
	var refreshID uuid.UUID
	if claims, ok := refreshToken.Claims.(jwt.MapClaims); ok && refreshToken.Valid {
		exp := claims["exp"].(float64)
		uuidID, err = uuid.Parse(claims["id"].(string))
		if err != nil {
			return uuid.Nil, false, fmt.Errorf("uuid.Parse - %w", err)
		}
		refreshID = uuidID
		if exp < float64(time.Now().Unix()) {
			return uuid.Nil, false, fmt.Errorf("validateToken - %w", err)
		}
	}
	if accessID != refreshID {
		return uuid.Nil, false, fmt.Errorf("user ID in acess token doesn't equal user ID in refresh token")
	}
	return accessID, isAdmin, nil
}

// HashPassword is a method of ServiceUser that makes from bytes hashed value
func (s *UserService) HashPassword(password []byte) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword(password, constants.BcryptCost)
	if err != nil {
		return bytes, fmt.Errorf("bcrypt.GenerateFromPassword - %w", err)
	}
	return bytes, nil
}

// CheckPasswordHash is a method of ServiceUser that checks if hash is equal hash from given password
func (s *UserService) CheckPasswordHash(hash, password []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, password)
	if err != nil {
		return false, fmt.Errorf("bcrypt.CompareHashAndPassword - %w", err)
	}
	return true, nil
}

// GenerateTokenPair generates pair of access and refresh tokens
func (s *UserService) GenerateTokenPair(id uuid.UUID, isAdmin bool) (TokenPair, error) {
	accessToken, err := s.GenerateJWTToken(constants.AccessTokenExpiration, id, isAdmin)
	if err != nil {
		return TokenPair{}, fmt.Errorf("GenerateJWTToken - %w", err)
	}
	refreshToken, err := s.GenerateJWTToken(constants.RefreshTokenExpiration, id, isAdmin)
	if err != nil {
		return TokenPair{}, fmt.Errorf("GenerateJWTToken - %w", err)
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateJWTToken is a method of ServiceUser that generate JWT token with given expiration with user id
func (s *UserService) GenerateJWTToken(expiration time.Duration, id uuid.UUID, isAdmin bool) (string, error) {
	claims := &jwt.MapClaims{
		"exp":     time.Now().Add(expiration).Unix(),
		"id":      id,
		"isAdmin": isAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.BlogTokenSignature))
	if err != nil {
		return "", fmt.Errorf("token.SignedString - %w", err)
	}
	return tokenString, nil
}
