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
func MapToUserListResponse(users []*domain.User, totalCount int64, page, pageSize int) UserListResponseDto {
	userResponses := make([]UserResponseDto, len(users))
	//mapping between layers
	for i, user := range users {
		userResponses[i] = MapToUserResponse(user)
	}

	//total pages
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize != 0 {
		totalPages++
	}

	return UserListResponseDto{
		Users:      userResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

}
