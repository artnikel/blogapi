package service

import (
	"context"

	"github.com/artnikel/blogapi/internal/model"
	"github.com/google/uuid"
)

type BlogRepository interface {
	Create(ctx context.Context, blog *model.Blog) error
	Get(ctx context.Context, id uuid.UUID) (*model.Blog, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByProfileID(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, blog *model.Blog) error
	GetAll(ctx context.Context) ([]*model.Blog, error)
	GetByProfileID(ctx context.Context, id uuid.UUID) ([]*model.Blog, error)
}

type BlogService struct {
	blogRps BlogRepository
}

func NewBlogService(blogRps BlogRepository) *BlogService {
	return &BlogService{blogRps: blogRps}
}
