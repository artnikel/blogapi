package service

import (
	"context"
	"fmt"

	"github.com/artnikel/blogapi/internal/model"
	"github.com/google/uuid"
)

type BlogRepository interface {
	Create(ctx context.Context, blog *model.Blog) error
	Get(ctx context.Context, id uuid.UUID) (*model.Blog, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, blog *model.Blog) error
	GetAll(ctx context.Context) ([]*model.Blog, error)
	GetByUserID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error)
}

type BlogService struct {
	blogRps BlogRepository
}

func NewBlogService(blogRps BlogRepository) *BlogService {
	return &BlogService{blogRps: blogRps}
}

func (s *BlogService) Create(ctx context.Context, blog *model.Blog) error {
	err := s.blogRps.Create(ctx, blog)
	if err != nil {
		return fmt.Errorf("blogRps.Create - %w", err)
	}
	return nil
}

func (s *BlogService) Get(ctx context.Context, id uuid.UUID) (*model.Blog, error) {
	blog, err := s.blogRps.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("blogRps.Get - %w", err)
	}
	return blog, nil
}

func (s *BlogService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.blogRps.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("blogRps.Delete - %w", err)
	}
	return nil
}

func (s *BlogService) DeleteByUserID(ctx context.Context, id uuid.UUID) error {
	err := s.blogRps.DeleteByUserID(ctx, id)
	if err != nil {
		return fmt.Errorf("blogRps.DeleteByUserID - %w", err)
	}
	return nil
}

func (s *BlogService) Update(ctx context.Context, blog *model.Blog) error {
	err := s.blogRps.Update(ctx, blog)
	if err != nil {
		return fmt.Errorf("blogRps.Update - %w", err)
	}
	return nil
}

func (s *BlogService) GetAll(ctx context.Context) ([]*model.Blog, error) {
	blogs, err := s.blogRps.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("blogRps.GetAll - %w", err)
	}
	return blogs, nil
}

func (s *BlogService) GetByUserID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error) {
	blogs, err := s.blogRps.GetByUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("blogRps.GetByUserID - %w", err)
	}
	return blogs, nil
}
