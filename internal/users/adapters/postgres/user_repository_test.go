package postgres

import (
	"testing"

	"github.com/cristianortiz/observ-monit-go/internal/users/domain"
	"github.com/stretchr/testify/assert"
)

// setupTestUser crea un user de prueba
func setupTestUser() *domain.User {
	user, _ := domain.NewUser(
		"John Doe",
		"john@example.com",
		"SecurePass123!",
	)
	return user
}

// TestUserRepository_Create tests the happy path and duplicate email
func TestUserRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// TODO: Aquí usaremos testcontainers o mock
	// Por ahora, estructura básica
	t.Run("success - creates user", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)
		// user := setupTestUser()

		// err := repo.Create(ctx, user)
		// require.NoError(t, err)

		// // Verify it was created
		// found, err := repo.GetByID(ctx, user.ID)
		// require.NoError(t, err)
		// assert.Equal(t, user.Email, found.Email)
	})

	t.Run("error - duplicate email", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)
		// user := setupTestUser()

		// // Create first time
		// err := repo.Create(ctx, user)
		// require.NoError(t, err)

		// // Try to create again with same email
		// user2, _ := domain.NewUser("Jane Doe", user.Email, "Pass123!")
		// err = repo.Create(ctx, user2)

		// assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})
}

// TestUserRepository_GetByID tests finding and not finding users
func TestUserRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("success - finds existing user", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)
		// user := setupTestUser()

		// // Create user first
		// err := repo.Create(ctx, user)
		// require.NoError(t, err)

		// // Get by ID
		// found, err := repo.GetByID(ctx, user.ID)
		// require.NoError(t, err)
		// assert.Equal(t, user.ID, found.ID)
		// assert.Equal(t, user.Email, found.Email)
	})

	t.Run("error - user not found", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)

		// nonExistentID := uuid.New().String()
		// _, err := repo.GetByID(ctx, nonExistentID)

		// assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

// TestUserRepository_GetByEmail tests email lookup
func TestUserRepository_GetByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("success - finds user by email", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)
		// user := setupTestUser()

		// err := repo.Create(ctx, user)
		// require.NoError(t, err)

		// found, err := repo.GetByEmail(ctx, user.Email)
		// require.NoError(t, err)
		// assert.Equal(t, user.ID, found.ID)
	})

	t.Run("error - email not found", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)

		// _, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		// assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

// TestUserRepository_Update tests update scenarios
func TestUserRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("success - updates user", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)
		// user := setupTestUser()

		// // Create
		// err := repo.Create(ctx, user)
		// require.NoError(t, err)

		// // Update
		// user.Name = "Jane Doe"
		// user.UpdatedAt = time.Now()
		// err = repo.Update(ctx, user)
		// require.NoError(t, err)

		// // Verify
		// found, err := repo.GetByID(ctx, user.ID)
		// require.NoError(t, err)
		// assert.Equal(t, "Jane Doe", found.Name)
	})

	t.Run("error - user not found", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)
		// user := setupTestUser()

		// err := repo.Update(ctx, user)
		// assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("error - duplicate email on update", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)

		// user1 := setupTestUser()
		// user2, _ := domain.NewUser("Jane Doe", "jane@example.com", "Pass123!")

		// // Create both
		// require.NoError(t, repo.Create(ctx, user1))
		// require.NoError(t, repo.Create(ctx, user2))

		// // Try to update user2 with user1's email
		// user2.Email = user1.Email
		// err := repo.Update(ctx, user2)

		// assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})
}

// TestUserRepository_Delete tests deletion
func TestUserRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("success - deletes user", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)
		// user := setupTestUser()

		// // Create
		// err := repo.Create(ctx, user)
		// require.NoError(t, err)

		// // Delete
		// err = repo.Delete(ctx, user.ID)
		// require.NoError(t, err)

		// // Verify it's gone
		// _, err = repo.GetByID(ctx, user.ID)
		// assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("error - user not found", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)

		// nonExistentID := uuid.New().String()
		// err := repo.Delete(ctx, nonExistentID)

		// assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

// TestUserRepository_List tests pagination
func TestUserRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("success - lists users with pagination", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)

		// // Create 3 users
		// for i := 0; i < 3; i++ {
		// 	user, _ := domain.NewUser(
		// 		fmt.Sprintf("User %d", i),
		// 		fmt.Sprintf("user%d@example.com", i),
		// 		"Pass123!",
		// 	)
		// 	require.NoError(t, repo.Create(ctx, user))
		// }

		// // List with pagination
		// users, err := repo.List(ctx, 10, 0)
		// require.NoError(t, err)
		// assert.GreaterOrEqual(t, len(users), 3)
	})

	t.Run("success - returns empty list when no users", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)

		// users, err := repo.List(ctx, 10, 0)
		// require.NoError(t, err)
		// assert.Empty(t, users)
	})
}

// TestUserRepository_Count tests counting
func TestUserRepository_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("success - counts users", func(t *testing.T) {
		t.Skip("pending database setup")

		// ctx := context.Background()
		// repo := NewUserRepository(testDB)

		// initialCount, err := repo.Count(ctx)
		// require.NoError(t, err)

		// // Create a user
		// user := setupTestUser()
		// require.NoError(t, repo.Create(ctx, user))

		// // Count again
		// newCount, err := repo.Count(ctx)
		// require.NoError(t, err)
		// assert.Equal(t, initialCount+1, newCount)
	})
}

// Helper: isUniqueViolation test (unit test - no DB needed)
func TestIsUniqueViolation(t *testing.T) {
	t.Run("returns false for nil error", func(t *testing.T) {
		result := isUniqueViolation(nil, constraintUsersEmailKey)
		assert.False(t, result)
	})

	t.Run("returns false for non-pg error", func(t *testing.T) {
		err := assert.AnError
		result := isUniqueViolation(err, constraintUsersEmailKey)
		assert.False(t, result)
	})

	// Note: Testing true case requires a real PgError
	// which is hard to mock, so we skip it (integration test will cover it)
}
