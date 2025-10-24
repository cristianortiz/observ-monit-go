package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/cristianortiz/observ-monit-go/internal/users/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL error codes that maps to domain errors
const (
	pgCodeUniqueViolation = "23505"
)

// Constraint names del schema
const (
	constraintUsersEmailKey = "users_email_key"
)

// UserRepository implements domain.UserRepository using  PostgreSQL
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository crea una nueva instancia del repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// duplicated email is a domain error
		if isUniqueViolation(err, constraintUsersEmailKey) {
			return domain.ErrEmailAlreadyExists
		}

		// any other error from postgres is a generic one
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
        SELECT id, name, email, password_hash, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT id, name, email, password_hash, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
        UPDATE users
        SET name = $2, email = $3, password_hash = $4, updated_at = $5
        WHERE id = $1
    `

	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err, constraintUsersEmailKey) {
			return domain.ErrEmailAlreadyExists
		}

		return fmt.Errorf("failed to update user: %w", err)
	}

	// ✅ Verificar que se actualizó al menos 1 fila
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		// ❌ No hay errores de dominio específicos para delete
		// (podrías agregar foreign key violation si es necesario)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// ✅ Verificar que se eliminó al menos 1 fila
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
        SELECT id, name, email, password_hash, created_at, updated_at
        FROM users
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		//  Error genérico (no es de dominio)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// ============================================================
// HELPERS - Funciones auxiliares privadas
// ============================================================

// isUniqueViolation checks if the errir is unique constrain for a field, like email
func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgCodeUniqueViolation &&
			pgErr.ConstraintName == constraintName
	}
	return false
}
