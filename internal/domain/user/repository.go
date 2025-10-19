package user

import "context"

// Repository : Contract to access the user persistence layer
// this is a CONTRACT, the real implementation will be in internal/infrastructure/persistence
// interface approach reasoning
// 1. decoupling domain layer from implmentation (Postgres, mongoDB, etc)
// 2. Easy testing, mocks to simulate persistence layer
// 3. Follows the Dependency Inversion Principle
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*User, error)
}
