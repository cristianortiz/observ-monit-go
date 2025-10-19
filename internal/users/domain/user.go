package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User domain entity, represents a system user wih their bussiness rules
type User struct {
	ID           string // UUID v4
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time // Pointer =nullable (soft delete)

}

func NewUser(name, email, password string) (*User, error) {
	// first user bussines rules: normalize email
	email = strings.ToLower(strings.TrimSpace(email))
	//second bussines rule: normalize name
	name = strings.TrimSpace(name)
	//third bussines rule: hashing password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	//create user with init values
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
	}, nil

}

// ValidatePassword checks if password is correct, will be used by login or auth service
// - returns true if password is valid, the plain text correspond to  stored hash
func (u *User) ValidatePassword(p string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash),
		[]byte(p),
	)
	return err == nil

}

// IsDeleted checks if the user was soft-deleted
// the soft delete flag the suer as deleted but withoud remove it phisically fromm DB, this enables
// - Auditory: to know when was "deletef"
// - Recovery: be abble to restore the user if delete op was an error
// - Integrity : Does not breaks the foreign key dependency
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// SoftDelete: flasg the user as "deleted", updates UpdateAt to reflex the change
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
	u.UpdatedAt = now
}
