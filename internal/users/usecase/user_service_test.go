package usecase

import (
	"context"
	"testing"

	"github.com/cristianortiz/observ-monit-go/internal/users/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of domain.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return int64(args.Int(0)), args.Error(1)
}

// cloneUser creates a copy of a user (to avoid reference issues in tests)
func cloneUser(u *domain.User) *domain.User {
	clone := *u
	return &clone
}

// ============================================================
// TESTS
// ============================================================

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success - creates user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		// Mock: email doesn't exist
		mockRepo.On("GetByEmail", ctx, "john@example.com").
			Return(nil, domain.ErrUserNotFound)

		// Mock: create succeeds
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).
			Return(nil)

		user, err := service.CreateUser(ctx, "John Doe", "john@example.com", "SecurePass123!")

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - email already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		existingUser, _ := domain.NewUser("Jane Doe", "john@example.com", "Pass123!")

		// Mock: email exists
		mockRepo.On("GetByEmail", ctx, "john@example.com").
			Return(existingUser, nil)

		user, err := service.CreateUser(ctx, "John Doe", "john@example.com", "SecurePass123!")

		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

}

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success - finds user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		expectedUser, _ := domain.NewUser("John Doe", "john@example.com", "Pass123!")

		mockRepo.On("GetByID", ctx, expectedUser.ID).
			Return(expectedUser, nil)

		user, err := service.GetUserByID(ctx, expectedUser.ID)

		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent-id").
			Return(nil, domain.ErrUserNotFound)

		user, err := service.GetUserByID(ctx, "non-existent-id")

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success - updates user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		existingUser, _ := domain.NewUser("John Doe", "john@example.com", "Pass123!")

		// Mock: get existing user
		mockRepo.On("GetByID", ctx, existingUser.ID).
			Return(cloneUser(existingUser), nil)

		// Mock: new email doesn't exist
		mockRepo.On("GetByEmail", ctx, "newemail@example.com").
			Return(nil, domain.ErrUserNotFound)

		// Mock: update succeeds
		mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
			Return(nil)

		user, err := service.UpdateUser(ctx, existingUser.ID, "Jane Doe", "newemail@example.com")

		require.NoError(t, err)
		assert.Equal(t, "Jane Doe", user.Name)
		assert.Equal(t, "newemail@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent-id").
			Return(nil, domain.ErrUserNotFound)

		user, err := service.UpdateUser(ctx, "non-existent-id", "Jane Doe", "jane@example.com")

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - email already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		existingUser, _ := domain.NewUser("John Doe", "john@example.com", "Pass123!")
		otherUser, _ := domain.NewUser("Jane Doe", "jane@example.com", "Pass123!")

		mockRepo.On("GetByID", ctx, existingUser.ID).
			Return(cloneUser(existingUser), nil)

		mockRepo.On("GetByEmail", ctx, "jane@example.com").
			Return(otherUser, nil)

		user, err := service.UpdateUser(ctx, existingUser.ID, "John Doe", "jane@example.com")

		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success - deletes user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		existingUser, _ := domain.NewUser("John Doe", "john@example.com", "Pass123!")

		mockRepo.On("GetByID", ctx, existingUser.ID).
			Return(existingUser, nil)

		mockRepo.On("Delete", ctx, existingUser.ID).
			Return(nil)

		err := service.DeleteUser(ctx, existingUser.ID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - user not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		mockRepo.On("GetByID", ctx, "non-existent-id").
			Return(nil, domain.ErrUserNotFound)

		err := service.DeleteUser(ctx, "non-existent-id")

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("success - lists users with pagination", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		user1, _ := domain.NewUser("User 1", "user1@example.com", "Pass123!")
		user2, _ := domain.NewUser("User 2", "user2@example.com", "Pass123!")
		expectedUsers := []*domain.User{user1, user2}

		mockRepo.On("List", ctx, 20, 0).
			Return(expectedUsers, nil)

		mockRepo.On("Count", ctx).
			Return(2, nil)

		users, total, err := service.ListUsers(ctx, 20, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, 2, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - applies default pagination", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)

		mockRepo.On("List", ctx, 20, 0).
			Return([]*domain.User{}, nil)

		mockRepo.On("Count", ctx).
			Return(0, nil)

		users, total, err := service.ListUsers(ctx, 0, -1) // Invalid params

		require.NoError(t, err)
		assert.Equal(t, 0, len(users))
		assert.Equal(t, 0, total)
		mockRepo.AssertExpectations(t)
	})
}
