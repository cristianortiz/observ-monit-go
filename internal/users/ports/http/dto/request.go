package dto

type CreateUserRequestDto struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequestDto struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

type UpdatePasswordRequestDto struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type LoginRequestDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ListUsersQueryDto struct {
	Limit  int `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

func (q *ListUsersQueryDto) SetDefaults() {
	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.Limit > 100 { //límite máximo
		q.Limit = 100
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
}
