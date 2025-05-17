// Package model provides the data models used in the application
package model

import (
	"time"

	"github.com/google/uuid"
)

// Blog entity
type Blog struct {
	BlogID      uuid.UUID `json:"blogid,omitempty" validate:"required"`
	UserID      uuid.UUID `json:"userid,omitempty"`
	Title       string    `json:"title" validate:"required"`
	Content     string    `json:"content" validate:"required"`
	ReleaseTime time.Time `json:"releasetime"`
}

// User entity
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username" validate:"required,min=4,max=15"`
	Password     []byte    `json:"password" validate:"required,min=4,max=15"`
	RefreshToken string    `json:"refreshToken"`
	Admin        bool      `json:"-"`
}

// BlogListResponse is struct for pagination
type BlogListResponse struct {
	Blogs []*Blog `json:"blogs"`
	Count int     `json:"count"`
}
