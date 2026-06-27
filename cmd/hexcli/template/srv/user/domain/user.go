package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewUser(id, username, password string) (*User, error) {
	hashed, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:           id,
		Username:     username,
		PasswordHash: hashed,
		CreatedAt:    time.Now(),
	}, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
