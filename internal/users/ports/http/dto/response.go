package dto

import "time"

// UserResponseDto to return info about user, no password considered by security
type UserResponseDto struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponseDto struct {
	Users      []UserResponseDto `json:"users"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type LoginResponseDto struct {
	User  UserResponseDto `json:"user"`
	Token string          `json:"token,omitempty"` // JWT token
}

type ErrorResponseDto struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// for succes operations, like a delete op
type MessageResponse struct {
	Message string `json:"message"`
}
