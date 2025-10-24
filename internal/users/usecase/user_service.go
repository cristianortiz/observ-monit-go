package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/cristianortiz/observ-monit-go/internal/users/domain"
)

// UserService handles user business logic
type UserService struct {
	repo domain.UserRepository
}

// NewUserService creates a new user service instance
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// CreateUser creates a new user with validation
func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	// 1. Validate email uniqueness (business rule)
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}

	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// 2. Create domain entity (includes validation + password hashing)
	user, err := domain.NewUser(name, email, password)
	if err != nil {
		return nil, err // domain validation error
	}

	// 3. Persist to repository
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id, name, email string) (*domain.User, error) {
	// 1. Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Check if email is being changed to an existing one
	if user.Email != email {
		existing, err := s.repo.GetByEmail(ctx, email)
		if err != nil && err != domain.ErrUserNotFound {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}

		if existing != nil && existing.ID != id {
			return nil, domain.ErrEmailAlreadyExists
		}
	}

	// 3. Update fields
	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now()

	// 4. Persist changes
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	// Optional: verify user exists before attempting delete
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves paginated users
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 20 // default
	}

	if offset < 0 {
		offset = 0
	}

	// Get users
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count for pagination metadata
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

// // AuthenticateUser validates user credentials
// func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
// 	// 1. Get user by email
// 	user, err := s.repo.GetByEmail(ctx, email)
// 	if err != nil {
// 		if err == domain.ErrUserNotFound {
// 			return nil, domain.ErrInvalidCredentials
// 		}
// 		return nil, fmt.Errorf("failed to authenticate: %w", err)
// 	}

// 	return user, nil
// }
