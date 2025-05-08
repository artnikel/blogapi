package model

import (
	"time"

	"github.com/google/uuid"
)

type Blog struct {
	BlogID      uuid.UUID `json:"blogid,omitempty" validate:"required"`
	UserID   uuid.UUID `json:"userid,omitempty" validate:"required"`
	Title       string    `json:"title" validate:"required"`
	Content     string    `json:"content" validate:"required"`
	ReleaseTime time.Time `json:"releasetime"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username" validate:"required,min=4,max=15"`
	Password     []byte    `json:"password" validate:"required,min=4,max=15"`
	RefreshToken string    `json:"refreshToken"`
	Admin        bool      `json:"-"`
}
