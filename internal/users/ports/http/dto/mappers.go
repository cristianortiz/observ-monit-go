package dto

import "github.com/cristianortiz/observ-monit-go/internal/users/domain"

// mapToUserResponse converts domain.User to UserResponseDto, ready to serialize to JSON
func MapToUserResponse(user *domain.User) UserResponseDto {
	return UserResponseDto{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// MapToUserListReponse converts a slice of domain.User to UserListResponseDto with pagination details
func MapToUserListResponse(users []*domain.User, total int64, limit, offset int) UserListResponseDto {
	userResponses := make([]UserResponseDto, len(users))
	//mapping between layers
	for i, user := range users {
		userResponses[i] = MapToUserResponse(user)
	}

	// ✅ SAFE GUARD: Prevenir divide by zero
	if limit <= 0 {
		limit = 20 // Default fallback
	}
	if offset < 0 {
		offset = 0 // Default fallback
	}

	// Calcular totalPages de forma segura
	totalPages := 0
	if total > 0 && limit > 0 {
		totalPages = (int(total) + limit - 1) / limit // Ceiling division
	}

	// Calcular currentPage de forma segura
	currentPage := 1
	if limit > 0 {
		currentPage = (offset / limit) + 1
	}

	return UserListResponseDto{
		Users:      userResponses,
		TotalCount: total,
		PageSize:   limit,
		//Offset:     offset,
		TotalPages: totalPages,
		Page:       currentPage,
	}

}
