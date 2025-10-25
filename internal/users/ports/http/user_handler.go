package http

import (
	"errors"

	"github.com/cristianortiz/observ-monit-go/internal/users/domain"
	"github.com/cristianortiz/observ-monit-go/internal/users/ports/http/dto"
	"github.com/cristianortiz/observ-monit-go/internal/users/usecase"
	"github.com/gofiber/fiber/v2"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	service *usecase.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *usecase.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// CreateUser handles POST /api/users
// @Summary Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequestDto true "User creation request"
// @Success 201 {object} dto.UserResponseDto
// @Failure 400 {object} dto.ErrorResponseDto
// @Failure 409 {object} dto.ErrorResponseDto
// @Failure 500 {object} dto.ErrorResponseDto
// @Router /api/users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// Get validated data from middleware
	req := c.Locals("validated_data").(dto.CreateUserRequestDto)

	// Call service
	user, err := h.service.CreateUser(
		c.Context(),
		req.Name,
		req.Email,
		req.Password,
	)

	if err != nil {
		return h.handleError(c, err)
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(dto.MapToUserResponse(user))
}

// GetUser handles GET /api/users/:id
// @Summary Get user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponseDto
// @Failure 404 {object} dto.ErrorResponseDto
// @Failure 500 {object} dto.ErrorResponseDto
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := h.service.GetUserByID(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(dto.MapToUserResponse(user))
}

// UpdateUser handles PUT /api/users/:id
// @Summary Update user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequestDto true "User update request"
// @Success 200 {object} dto.UserResponseDto
// @Failure 400 {object} dto.ErrorResponseDto
// @Failure 404 {object} dto.ErrorResponseDto
// @Failure 409 {object} dto.ErrorResponseDto
// @Failure 500 {object} dto.ErrorResponseDto
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get validated data from middleware
	req := c.Locals("validated_data").(dto.UpdateUserRequestDto)

	// Call service
	user, err := h.service.UpdateUser(
		c.Context(),
		id,
		*req.Name,
		*req.Email,
	)

	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(dto.MapToUserResponse(user))
}

// DeleteUser handles DELETE /api/users/:id
// @Summary Delete user
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} dto.ErrorResponseDto
// @Failure 500 {object} dto.ErrorResponseDto
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.service.DeleteUser(c.Context(), id)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsers handles GET /api/users
// @Summary List users with pagination
// @Tags users
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.UserListResponseDto
// @Failure 400 {object} dto.ErrorResponseDto
// @Failure 500 {object} dto.ErrorResponseDto
// @Router /api/users [get]
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	// Get validated query params from middleware
	query := c.Locals("validated_query").(dto.ListUsersQueryDto)

	// Call service
	users, total, err := h.service.ListUsers(c.Context(), query.Limit, query.Offset)
	if err != nil {
		return h.handleError(c, err)
	}

	// Build response
	response := dto.MapToUserListResponse(users, total, query.Limit, query.Offset)

	return c.Status(fiber.StatusOK).JSON(response)
}

// handleError maps domain errors to HTTP responses
func (h *UserHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponseDto{
			Error:   "Not Found",
			Message: "User not found",
		})

	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponseDto{
			Error:   "Conflict",
			Message: "Email already exists",
		})

	// case errors.Is(err, domain.ErrInvalidUserData):
	// 	return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponseDto{
	// 		Error:   "Bad Request",
	// 		Message: err.Error(),
	// 	})

	case errors.Is(err, domain.ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponseDto{
			Error:   "Unauthorized",
			Message: "Invalid credentials",
		})

	default:
		// Log internal error (TODO: add logger)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponseDto{
			Error:   "Internal Server Error",
			Message: "An unexpected error occurred",
		})
	}
}
