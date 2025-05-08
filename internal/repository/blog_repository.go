package repository

import (
	"context"
	"fmt"

	"github.com/artnikel/blogapi/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{
		pool: pool,
	}
}

func (p *PgRepository) Create(ctx context.Context, blog *model.Blog) error {
	_, err := p.pool.Exec(ctx, "INSERT INTO blog (blogid, profileid, title, content) VALUES ($1, $2, $3)",
		blog.BlogID, blog.ProfileID, blog.Title, blog.Content)
	if err != nil {
		return fmt.Errorf("PgRepository-Create: error in method p.pool.Exec(): %w", err)
	}
	return nil
}

func (p *PgRepository) Get(ctx context.Context, id uuid.UUID) (*model.Blog, error) {
	var blog model.Blog
	err := p.pool.QueryRow(ctx, "SELECT blogid, profileid, title, content, releasetime FROM blog WHERE blogid = $1", id).
		Scan(&blog.BlogID, &blog.ProfileID, &blog.Title, &blog.Content, &blog.ReleaseTime)
	if err != nil {
		return nil, fmt.Errorf("PgRepository-Get: error in method p.pool.QuerryRow(): %w", err)
	}
	return &blog, nil
}

func (p *PgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM blog WHERE blogid = $1", id)
	if err != nil {
		return fmt.Errorf("PgRepository-Delete: error in method p.pool.Exec(): %w", err)
	}
	return nil
}

func (p *PgRepository) DeleteByProfileID(ctx context.Context, id uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM blog WHERE profileid = $1", id)
	if err != nil {
		return fmt.Errorf("PgRepository-DeleteByProfileID: error in method p.pool.Exec(): %w", err)
	}
	return nil
}

func (p *PgRepository) Update(ctx context.Context, blog *model.Blog) error {
	_, err := p.pool.Exec(ctx, "UPDATE blog SET title = $1, content = $2 WHERE blogid = $3", blog.Title, blog.Content, blog.BlogID)
	if err != nil {
		return fmt.Errorf("PgRepository-Update: error in method p.pool.Exec(): %w", err)
	}
	return nil
}

func (p *PgRepository) GetAll(ctx context.Context) ([]*model.Blog, error) {
	var blogs []*model.Blog
	rows, err := p.pool.Query(ctx, "SELECT blogid, profileid, title, content, releasetime FROM blog")
	if err != nil {
		return nil, fmt.Errorf("PgRepository-GetAll: error in method p.pool.Query(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var blog model.Blog
		err := rows.Scan(&blog.BlogID, &blog.ProfileID, &blog.Title, &blog.Content, &blog.ReleaseTime)
		if err != nil {
			return nil, fmt.Errorf("PgRepository-GetAll: error in method rows.Scan(): %w", err)
		}
		blogs = append(blogs, &blog)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("PgRepository-GetAll: error iterating rows: %w", err)
	}
	return blogs, nil
}

func (p *PgRepository) GetByProfileID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error) {
	var blogs []*model.Blog
	rows, err := p.pool.Query(ctx, "SELECT profileid, blogid, title, content, releasetime FROM blog WHERE profileid = $1", id)
	if err != nil {
		return nil, fmt.Errorf("PgRepository-GetByProfileID: error in method p.pool.QuerryRow(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var blog model.Blog
		err := rows.Scan(&blog.ProfileID, &blog.BlogID, &blog.Title, &blog.Content, &blog.ReleaseTime)
		if err != nil {
			return nil, fmt.Errorf("PgRepository-GetByProfileID: error in method rows.Scan(): %w", err)
		}
		blogs = append(blogs, &blog)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("PgRepository-GetByProfileID: error iterating rows: %w", err)
	}
	return blogs, nil
}
