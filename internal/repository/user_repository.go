package repository

import (
	"context"
	"fmt"

	"github.com/artnikel/blogapi/internal/model"
	"github.com/google/uuid"
)

// SignUp creates a new user record in the db
func (p *PgRepository) SignUp(ctx context.Context, user *model.User) error {
	if user == nil {
		return ErrNil
	}
	var numberUsers int
	err := p.pool.QueryRow(context.Background(), "SELECT COUNT(id) FROM users WHERE username = $1", user.Username).Scan(&numberUsers)
	if err != nil {
		return fmt.Errorf("error in method p.pool.QueryRow(): %w", err)
	}
	if numberUsers != 0 {
		return ErrExist
	}
	_, err = p.pool.Exec(ctx, "INSERT INTO users(id, username, password, admin) VALUES($1, $2, $3, $4)",
		user.ID, user.Username, user.Password, user.Admin)
	if err != nil {
		return fmt.Errorf("error in method p.pool.Exec(): %w", err)
	}
	return nil
}

// GetDataByUsername returns data of user by username
func (p *PgRepository) GetDataByUsername(ctx context.Context, username string) (id uuid.UUID, password []byte, admin bool, e error) {
	var user model.User
	user.Username = username
	err := p.pool.QueryRow(ctx, "SELECT id, password, admin FROM users WHERE username = $1", user.Username).
		Scan(&user.ID, &user.Password, &user.Admin)
	if err != nil {
		return uuid.UUID{}, nil, false, fmt.Errorf("error in method p.pool.QueryRow(): %w", err)
	}
	return user.ID, user.Password, user.Admin, nil
}

// GetRefreshTokenByID returns refreshToken from users table by id
func (p *PgRepository) GetRefreshTokenByID(ctx context.Context, id uuid.UUID) (string, error) {
	var hash string
	err := p.pool.QueryRow(ctx, "SELECT refreshToken FROM users WHERE id = $1", id).Scan(&hash)
	if err != nil {
		return "", fmt.Errorf("error in method p.pool.QueryRow(): %w", err)
	}
	return hash, nil
}

// AddRefreshToken adds refreshToken to users table by id
func (p *PgRepository) AddRefreshToken(ctx context.Context, user *model.User) error {
	_, err := p.pool.Exec(ctx, "UPDATE users SET refreshtoken = $1 WHERE id = $2", user.RefreshToken, user.ID)
	if err != nil {
		return fmt.Errorf("error in method p.pool.Exec(): %w", err)
	}
	return nil
}
