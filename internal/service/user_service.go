// Package service realize bisnes-logic of the microservice
package service

import (
	"context"
	"crypto/sha256"
	"time"

	"fmt"

	"github.com/artnikel/blogapi/internal/config"
	"github.com/artnikel/blogapi/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	SignUp(ctx context.Context, user *model.User) error
	GetDataByUsername(ctx context.Context, username string) (uuid.UUID, []byte, bool, error)
	AddRefreshToken(ctx context.Context, user *model.User) error
	GetRefreshTokenByID(ctx context.Context, id uuid.UUID) (string, error)
}

type UserService struct {
	rpsUser UserRepository
	cfg     *config.Config
}

func NewUserService(rpsUser UserRepository, cfg *config.Config) *UserService {
	return &UserService{rpsUser: rpsUser, cfg: cfg}
}

const (
	accessTokenExpiration  = 15 * time.Minute
	refreshTokenExpiration = 72 * time.Hour
	bcryptCost             = 14
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (srvUser *UserService) SignUp(ctx context.Context, user *model.User) error {
	var err error
	user.Password, err = srvUser.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("HashPassword - %w", err)
	}
	err = srvUser.rpsUser.SignUp(ctx, user)
	if err != nil {
		return fmt.Errorf("SignUp - %w", err)
	}
	return nil
}

func (srvUser *UserService) Login(ctx context.Context, user *model.User) (*TokenPair, error) {
	id, hash, admin, err := srvUser.rpsUser.GetDataByUsername(ctx, user.Username)
	user.ID = id
	user.Admin = admin
	if err != nil {
		return &TokenPair{}, fmt.Errorf("GetDataByUsername - %w", err)
	}
	verified, err := srvUser.CheckPasswordHash(hash, user.Password)
	if err != nil || !verified {
		return &TokenPair{}, fmt.Errorf("CheckPasswordHash - %w", err)
	}
	tokenPair, err := srvUser.GenerateTokenPair(user.ID, user.Admin)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("GenerateTokenPair - %w", err)
	}
	sum := sha256.Sum256([]byte(tokenPair.RefreshToken))
	hashedRefreshToken, err := srvUser.HashPassword(sum[:])
	if err != nil {
		return &TokenPair{}, fmt.Errorf("HashPassword - %w", err)
	}
	user.RefreshToken = string(hashedRefreshToken)
	err = srvUser.rpsUser.AddRefreshToken(context.Background(), user)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("AddRefreshToken - %w", err)
	}
	return &tokenPair, nil
}

func (srvUser *UserService) Refresh(ctx context.Context, tokenPair TokenPair) (*TokenPair, error) {
	id, isAdmin, err := srvUser.TokensIDCompare(tokenPair)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("TokensIDCompare - %w", err)
	}
	hash, err := srvUser.rpsUser.GetRefreshTokenByID(ctx, id)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("GetRefreshTokenByID - %w", err)
	}
	sum := sha256.Sum256([]byte(tokenPair.RefreshToken))
	verified, err := srvUser.CheckPasswordHash([]byte(hash), sum[:])
	if err != nil || !verified {
		return &TokenPair{}, fmt.Errorf("CheckPasswordHash error: refreshToken invalid")
	}
	tokenPair, err = srvUser.GenerateTokenPair(id, isAdmin)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("GenerateTokenPair - %w", err)
	}
	sum = sha256.Sum256([]byte(tokenPair.RefreshToken))
	hashedRefreshToken, err := srvUser.HashPassword(sum[:])
	if err != nil {
		return &TokenPair{}, fmt.Errorf("HashPassword - %w", err)
	}
	var user model.User
	user.RefreshToken = string(hashedRefreshToken)
	user.ID = id
	err = srvUser.rpsUser.AddRefreshToken(context.Background(), &user)
	if err != nil {
		return &TokenPair{}, fmt.Errorf("AddRefreshToken - %w", err)
	}
	return &tokenPair, nil
}

func (srvUser *UserService) TokensIDCompare(tokenPair TokenPair) (uuid.UUID, bool, error) {
	accessToken, err := validateToken(tokenPair.AccessToken, srvUser.cfg.BlogTokenSignature)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("validateToken - %w", err)
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
	refreshToken, err := validateToken(tokenPair.RefreshToken, srvUser.cfg.BlogTokenSignature)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("validateToken - %w", err)
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

func (srvUser *UserService) HashPassword(password []byte) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword(password, bcryptCost)
	if err != nil {
		return bytes, fmt.Errorf("GenerateFromPassword - %w", err)
	}
	return bytes, nil
}

func (srvUser *UserService) CheckPasswordHash(hash, password []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, password)
	if err != nil {
		return false, fmt.Errorf("CompareHashAndPassword - %w", err)
	}
	return true, nil
}

func (srvUser *UserService) GenerateTokenPair(id uuid.UUID, isAdmin bool) (TokenPair, error) {
	accessToken, err := srvUser.GenerateJWTToken(accessTokenExpiration, id, isAdmin)
	if err != nil {
		return TokenPair{}, fmt.Errorf("GenerateJWTToken - %w", err)
	}
	refreshToken, err := srvUser.GenerateJWTToken(refreshTokenExpiration, id, isAdmin)
	if err != nil {
		return TokenPair{}, fmt.Errorf("GenerateJWTToken - %w", err)
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (srvUser *UserService) GenerateJWTToken(expiration time.Duration, id uuid.UUID, isAdmin bool) (string, error) {
	claims := &jwt.MapClaims{
		"exp":     time.Now().Add(expiration).Unix(),
		"id":      id,
		"isAdmin": isAdmin,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(srvUser.cfg.BlogTokenSignature))
	if err != nil {
		return "", fmt.Errorf("token.SignedString - %w", err)
	}
	return tokenString, nil
}

func validateToken(tokenString, secretKey string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

