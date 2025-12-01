package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/cristianortiz/observ-monit-go/internal/users/domain"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// serviceTracer es el tracer de OpenTelemetry para la capa de servicio/usecase
var serviceTracer = otel.Tracer("users.service")

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
	ctx, span := serviceTracer.Start(ctx, "UserService.CreateUser")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.email", email),
		attribute.String("user.name", name),
	)
	// Helper para attributes consistentes
	tracing.SetOperationAttributes(span, "CreateUser", tracing.OperationTypeCreate, tracing.EntityUser)

	// Validate email uniqueness (business rule)
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && err != domain.ErrUserNotFound {
		//error de negocio, registrar
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}

	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Create domain entity (includes validation + password hashing)
	user, err := domain.NewUser(name, email, password)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err // domain validation error
	}

	// Persist to repository
	if err := s.repo.Create(ctx, user); err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	span.SetAttributes(attribute.String("user.id", user.ID))
	tracing.RecordSuccess(span)
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, span := serviceTracer.Start(ctx, "UserService.GetUserByID")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", id))
	tracing.SetOperationAttributes(span, "GetUserByID", tracing.OperationTypeRead, tracing.EntityUser)

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {

		tracing.RecordError(span, err)
		return nil, err
	}
	span.SetAttributes(attribute.String("user.email", user.Email))
	tracing.RecordSuccess(span)

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := serviceTracer.Start(ctx, "UserService.GetUserByEmail")
	defer span.End()
	span.SetAttributes(attribute.String("user.email", email))
	tracing.SetOperationAttributes(span, "GetUserByEmail", tracing.OperationTypeRead, tracing.EntityUser)

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}
	span.SetAttributes(attribute.String("user.id", user.ID))
	tracing.RecordSuccess(span)
	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id, name, email string) (*domain.User, error) {
	ctx, span := serviceTracer.Start(ctx, "UserService.UpdateUser")
	defer span.End()
	span.SetAttributes(
		attribute.String("user.id", id),
		attribute.String("user.new_email", email),
		attribute.String("user.new_name", name),
	)
	tracing.SetOperationAttributes(span, "UpdateUser", tracing.OperationTypeUpdate, tracing.EntityUser)

	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}
	// Guardar email anterior para logging
	span.SetAttributes(attribute.String("user.old_email", user.Email))

	// Check if email is being changed to an existing one
	if user.Email != email {
		existing, err := s.repo.GetByEmail(ctx, email)
		if err != nil && err != domain.ErrUserNotFound {
			tracing.RecordError(span, err)
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}

		if existing != nil && existing.ID != id {
			tracing.RecordError(span, domain.ErrEmailAlreadyExists)
			return nil, domain.ErrEmailAlreadyExists
		}
	}

	// 3. Update fields
	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now()

	// 4. Persist changes
	if err := s.repo.Update(ctx, user); err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	tracing.RecordSuccess(span)
	return user, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	ctx, span := serviceTracer.Start(ctx, "UserService.DeleteUser")
	defer span.End()
	span.SetAttributes(attribute.String("user.id", id))
	tracing.SetOperationAttributes(span, "DeleteUser", tracing.OperationTypeDelete, tracing.EntityUser)

	// Optional: verify user exists before attempting delete
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		tracing.RecordError(span, err)
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	tracing.RecordSuccess(span)

	return nil
}

// ListUsers retrieves paginated users
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	ctx, span := serviceTracer.Start(ctx, "UserService.ListUsers")
	defer span.End()
	// Attributes de paginación
	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)
	tracing.SetOperationAttributes(span, "ListUsers", tracing.OperationTypeList, tracing.EntityUser)
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		limit = 20 // default
	}

	if offset < 0 {
		offset = 0
	}

	// Agregar valores normalizados si cambiaron
	span.SetAttributes(
		attribute.Int("pagination.limit_used", limit),
		attribute.Int("pagination.offset_used", offset),
	)

	// Get users
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count for pagination metadata
	total, err := s.repo.Count(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Agregar resultados
	span.SetAttributes(
		attribute.Int("result.count", len(users)),
		attribute.Int64("result.total", total),
	)

	tracing.RecordSuccess(span)
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
