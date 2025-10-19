package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewUser_Success:
func TestNewUser_Success(t *testing.T) {
	// Arrange teste data
	name := "John Doe"
	email := "john.doe@example.com"
	password := "password123"

	user, err := NewUser(name, email, password)
	//if require the tests stop, is useful for preconditions
	require.NoError(t, err, "NewUser must not return error")
	require.NotNil(t, user, "User must not be nil")

	assert.NotEmpty(t, user.ID, "ID must be genereated")
	assert.Len(t, user.ID, 36, "UUID must have 36 chars")

	// check basic data
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)

	assert.NotEmpty(t, user.PasswordHash, "Password must be hashed")
	assert.NotEqual(t, password, user.PasswordHash, "Hash must be different from password")
	assert.Contains(t, user.PasswordHash, "$2a$", "must be bcrypt hash")

	// Verificar timestamps
	assert.False(t, user.IsDeleted(), "User was deleted before, is not a new user")
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.Equal(t, user.CreatedAt, user.UpdatedAt, "CreatedAt == UpdatedAt must be equal for a new User")
}

func TestNewUser_EmailNormalization(t *testing.T) {
	tests := []struct {
		name          string
		inputEmail    string
		expectedEmail string
	}{
		{
			name:          "uppercase email",
			inputEmail:    "JOHN@EXAMPLE.COM",
			expectedEmail: "john@example.com",
		},
		{
			name:          "mixed case email",
			inputEmail:    "John.Doe@Example.COM",
			expectedEmail: "john.doe@example.com",
		},
		{
			name:          "email with spaces",
			inputEmail:    "  john@example.com  ",
			expectedEmail: "john@example.com",
		},
		{
			name:          "email already lowercase",
			inputEmail:    "john@example.com",
			expectedEmail: "john@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser("John Doe", tt.inputEmail, "password123")

			require.NoError(t, err)
			assert.Equal(t, tt.expectedEmail, user.Email,
				"email must be normalized,s lowercased and have no spaces")
		})
	}
}

func TestNewUser_NameNormalization(t *testing.T) {
	user, err := NewUser("  John Doe  ", "john@example.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, "John Doe", user.Name, "Extra spaces must be deleted")
}

func TestUser_ValidatePassword(t *testing.T) {
	password := "mySecurePassword123"
	user, err := NewUser("John Doe", "john@example.com", password)
	require.NoError(t, err)

	t.Run("correct password", func(t *testing.T) {
		valid := user.ValidatePassword(password)
		assert.True(t, valid, "Password correct")
	})

	t.Run("incorrect password", func(t *testing.T) {
		valid := user.ValidatePassword("wrongPassword")
		assert.False(t, valid, "incorrect password")
	})

	t.Run("empty password", func(t *testing.T) {
		valid := user.ValidatePassword("")
		assert.False(t, valid, "empty password invalid")
	})

	t.Run("case sensitive", func(t *testing.T) {
		user2, _ := NewUser("Jane", "jane@example.com", "Password123")

		assert.True(t, user2.ValidatePassword("Password123"))
		assert.False(t, user2.ValidatePassword("password123"))
		assert.False(t, user2.ValidatePassword("PASSWORD123"))
	})
}

// TestUser_IsDeleted: check soft delete
func TestUser_IsDeleted(t *testing.T) {
	user, err := NewUser("John Doe", "john@example.com", "password123")
	require.NoError(t, err)

	t.Run("initially not deleted", func(t *testing.T) {
		assert.False(t, user.IsDeleted(), "New User must not be flag as deleted")
		assert.Nil(t, user.DeletedAt, "DeletedAt must be nil")
	})

	// After soft delete
	t.Run("after soft delete", func(t *testing.T) {
		user.SoftDelete()

		assert.True(t, user.IsDeleted(), "User must be deleted")
		assert.NotNil(t, user.DeletedAt, "DeletedAt must not be nil")
		assert.WithinDuration(t, time.Now(), *user.DeletedAt, time.Second)
	})
}

// TestUser_SoftDelete: check that UpdatedAt aldo is updated
func TestUser_SoftDelete_UpdatesTimestamp(t *testing.T) {
	user, err := NewUser("John Doe", "john@example.com", "password123")
	require.NoError(t, err)

	oldUpdatedAt := user.UpdatedAt

	//wait  some time to see the difference in timestap
	time.Sleep(10 * time.Millisecond)

	user.SoftDelete()

	assert.True(t, user.UpdatedAt.After(oldUpdatedAt),
		"UpdatedAt must be updated after SoftDelete")
}

// TestUser_UpdateName: update Name
func TestUser_UpdateName(t *testing.T) {
	user, err := NewUser("John Doe", "john@example.com", "password123")
	require.NoError(t, err)

	oldUpdatedAt := user.UpdatedAt
	oldName := user.Name

	time.Sleep(10 * time.Millisecond)

	newName := "Jane Smith"
	user.UpdateName(newName)

	assert.Equal(t, newName, user.Name, "Name must be updated")
	assert.NotEqual(t, oldName, user.Name, "Old and new name must be different")
	assert.True(t, user.UpdatedAt.After(oldUpdatedAt),
		"UpdatedAt must be updated")
}

func TestUser_UpdateName_TrimsSpaces(t *testing.T) {
	user, err := NewUser("John Doe", "john@example.com", "password123")
	require.NoError(t, err)

	user.UpdateName("  Jane Smith  ")

	assert.Equal(t, "Jane Smith", user.Name,
		"spaces must be deleted when name is updated")
}

// TestUser_UpdatePassword: Change password
func TestUser_UpdatePassword(t *testing.T) {
	oldPassword := "oldPassword123"
	user, err := NewUser("John Doe", "john@example.com", oldPassword)
	require.NoError(t, err)

	oldHash := user.PasswordHash
	oldUpdatedAt := user.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	newPassword := "newSecurePassword456"
	err = user.UpdatePassword(newPassword)

	require.NoError(t, err, "UpdatePassword must not failed")
	assert.NotEqual(t, oldHash, user.PasswordHash,
		"Hash must change")
	assert.True(t, user.ValidatePassword(newPassword),
		"New password  must be valid")
	assert.False(t, user.ValidatePassword(oldPassword),
		"Old password must be Not valid")
	assert.True(t, user.UpdatedAt.After(oldUpdatedAt),
		"UpdatedAt must be updated")
}

// TestNewUser_PasswordHash_IsDifferent: every hash must be unique
func TestNewUser_PasswordHash_IsDifferent(t *testing.T) {
	password := "samePassword123"

	// Crear dos usuarios con mismo password
	user1, err1 := NewUser("John", "john@example.com", password)
	user2, err2 := NewUser("Jane", "jane@example.com", password)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Hashes Must Be DIFFERENTS (bcrypt use salt random)
	assert.NotEqual(t, user1.PasswordHash, user2.PasswordHash,
		"bcrypt must generate differents hashes even whit the same password")

	// But both password must be valid
	assert.True(t, user1.ValidatePassword(password))
	assert.True(t, user2.ValidatePassword(password))
}
