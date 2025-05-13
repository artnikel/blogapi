// Package repository provides a PostgreSQL querry`s implementation
package repository

import (
	"context"
	"fmt"

	"github.com/artnikel/blogapi/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepository represents the PostgreSQL repository implementation
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates and returns a new instance of PgRepository, using the provided pgxpool.Pool
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{
		pool: pool,
	}
}

// Create creates a new blog record in the db
func (p *PgRepository) Create(ctx context.Context, blog *model.Blog) error {
	_, err := p.pool.Exec(ctx, "INSERT INTO blog (blogid, userid, title, content) VALUES ($1, $2, $3, $4)",
		blog.BlogID, blog.UserID, blog.Title, blog.Content)
	if err != nil {
		return fmt.Errorf("error in method p.pool.Exec(): %w", err)
	}
	return nil
}

// Get retrieves a blog record from the db based on the provided ID
func (p *PgRepository) Get(ctx context.Context, id uuid.UUID) (*model.Blog, error) {
	var blog model.Blog
	err := p.pool.QueryRow(ctx, "SELECT blogid, userid, title, content, releasetime FROM blog WHERE blogid = $1", id).
		Scan(&blog.BlogID, &blog.UserID, &blog.Title, &blog.Content, &blog.ReleaseTime)
	if err != nil {
		return nil, fmt.Errorf("error in method p.pool.QuerryRow(): %w", err)
	}
	return &blog, nil
}

// Delete removes a blog record from the db based on the provided ID
func (p *PgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM blog WHERE blogid = $1", id)
	if err != nil {
		return fmt.Errorf("error in method p.pool.Exec(): %w", err)
	}
	return nil
}

// DeleteByUserID removes blog records from the db based on the user ID
func (p *PgRepository) DeleteByUserID(ctx context.Context, id uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM blog WHERE userid = $1", id)
	if err != nil {
		return fmt.Errorf("error in method p.pool.Exec(): %w", err)
	}
	return nil
}

// Update updates a blog record in the db
func (p *PgRepository) Update(ctx context.Context, blog *model.Blog) error {
	_, err := p.pool.Exec(ctx, "UPDATE blog SET title = $1, content = $2 WHERE blogid = $3", blog.Title, blog.Content, blog.BlogID)
	if err != nil {
		return fmt.Errorf("error in method p.pool.Exec(): %w", err)
	}
	return nil
}

// GetAll retrieves all blogs records from the db
func (p *PgRepository) GetAll(ctx context.Context) ([]*model.Blog, error) {
	var blogs []*model.Blog
	rows, err := p.pool.Query(ctx, "SELECT blogid, userid, title, content, releasetime FROM blog")
	if err != nil {
		return nil, fmt.Errorf("error in method p.pool.Query(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var blog model.Blog
		err := rows.Scan(&blog.BlogID, &blog.UserID, &blog.Title, &blog.Content, &blog.ReleaseTime)
		if err != nil {
			return nil, fmt.Errorf("error in method rows.Scan(): %w", err)
		}
		blogs = append(blogs, &blog)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return blogs, nil
}

// GetByUserID retrieves all blogs from the db of a certain user
func (p *PgRepository) GetByUserID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error) {
	var blogs []*model.Blog
	rows, err := p.pool.Query(ctx, "SELECT userid, blogid, title, content, releasetime FROM blog WHERE userid = $1", id)
	if err != nil {
		return nil, fmt.Errorf("error in method p.pool.QuerryRow(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var blog model.Blog
		err := rows.Scan(&blog.UserID, &blog.BlogID, &blog.Title, &blog.Content, &blog.ReleaseTime)
		if err != nil {
			return nil, fmt.Errorf("error in method rows.Scan(): %w", err)
		}
		blogs = append(blogs, &blog)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return blogs, nil
}
