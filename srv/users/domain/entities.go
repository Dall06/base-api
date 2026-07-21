package domain

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID           string    `json:"id" bun:"id,pk,type:uuid"`
	Email        string    `json:"email" bun:"email,notnull"`
	PasswordHash string    `json:"-" bun:"password_hash,notnull"`
	Name         string    `json:"name" bun:"name,notnull"`
	CreatedAt    time.Time `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
}

