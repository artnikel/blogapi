// Package service provides the business logic for blog application
package service

import (
	"context"
	"fmt"

	"github.com/artnikel/blogapi/internal/model"
	"github.com/google/uuid"
)

// BlogRepository is an interface that contains CRUD methods
type BlogRepository interface {
	Create(ctx context.Context, blog *model.Blog) error
	Get(ctx context.Context, id uuid.UUID) (*model.Blog, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteBlogsByUserID(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, blog *model.Blog) error
	Count(ctx context.Context) (int, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Blog, error)
	GetByUserID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error)
}

// BlogService contains Repository interface
type BlogService struct {
	blogRps BlogRepository
}

// NewBlogService accepts Repository object and returns an object of type *BlogService
func NewBlogService(blogRps BlogRepository) *BlogService {
	return &BlogService{blogRps: blogRps}
}

// Create is a method of BlogService that calls Create method of Repository
func (s *BlogService) Create(ctx context.Context, blog *model.Blog) error {
	err := s.blogRps.Create(ctx, blog)
	if err != nil {
		return fmt.Errorf("blogRps.Create - %w", err)
	}
	return nil
}

// Get is a method of BlogService that calls Get method of Repository
func (s *BlogService) Get(ctx context.Context, id uuid.UUID) (*model.Blog, error) {
	blog, err := s.blogRps.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("blogRps.Get - %w", err)
	}
	return blog, nil
}

// Delete is a method of BlogService that calls Delete method of Repository
func (s *BlogService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.blogRps.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("blogRps.Delete - %w", err)
	}
	return nil
}

// DeleteBlogsByUserID is a method of BlogService that calls DeleteBlogsByUserID method of Repository
func (s *BlogService) DeleteBlogsByUserID(ctx context.Context, id uuid.UUID) error {
	err := s.blogRps.DeleteBlogsByUserID(ctx, id)
	if err != nil {
		return fmt.Errorf("blogRps.DeleteBlogsByUserID - %w", err)
	}
	return nil
}

// Update is a method of BlogService that calls Update method of Repository
func (s *BlogService) Update(ctx context.Context, blog *model.Blog) error {
	err := s.blogRps.Update(ctx, blog)
	if err != nil {
		return fmt.Errorf("blogRps.Update - %w", err)
	}
	return nil
}

// GetAll is a method of BlogService that calls GetAll method of Repository
func (s *BlogService) GetAll(ctx context.Context, limit, offset int) (*model.BlogListResponse, error) {
	count, err := s.blogRps.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("blogRps.Count - %w", err)
	}

	blogs, err := s.blogRps.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("blogRps.GetAll - %w", err)
	}

	return &model.BlogListResponse{
		Blogs: blogs,
		Count: count,
	}, nil
}

// GetByUserID is a method of BlogService that calls GetByUserID method of Repository
func (s *BlogService) GetByUserID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error) {
	blogs, err := s.blogRps.GetByUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("blogRps.GetByUserID - %w", err)
	}
	return blogs, nil
}
