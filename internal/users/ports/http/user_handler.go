package http

import (
	"errors"

	"github.com/cristianortiz/observ-monit-go/internal/users/domain"
	"github.com/cristianortiz/observ-monit-go/internal/users/ports/http/dto"
	"github.com/cristianortiz/observ-monit-go/internal/users/usecase"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/logger"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/metrics"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	service *usecase.UserService
	metrics metrics.UserMetrics
	logger  *logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *usecase.UserService, metrics *metrics.UserMetrics, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		metrics: *metrics,
		logger:  logger,
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
	// 1. Enrich logger with trace context
	logger := h.logger.WithTraceContext(c.UserContext())

	// 2. Get validated data from middleware (already parsed and validated)
	req := c.Locals("validated_data").(dto.CreateUserRequestDto)

	// 3. Call service with UserContext (contains OpenTelemetry span)
	logger.Info("Creating user", zap.String("email", req.Email))
	user, err := h.service.CreateUser(
		c.UserContext(),
		req.Name,
		req.Email,
		req.Password,
	)

	if err != nil {
		logger.Error("Failed to create user",
			zap.String("email", req.Email),
			zap.Error(err))
		return h.handleError(c, err)
	}

	// 4. Increment metrics
	h.metrics.UsersCreated.Inc()

	// 5. Log success with trace correlation
	logger.Info("User created successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))

	// 6. Return response
	response := dto.MapToUserResponse(user)
	return c.Status(fiber.StatusCreated).JSON(response)
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
	// Enrich logger with trace context
	logger := h.logger.WithTraceContext(c.UserContext())

	id := c.Params("id")
	logger.Info("Getting user", zap.String("user_id", id))

	user, err := h.service.GetUserByID(c.UserContext(), id)
	if err != nil {
		logger.Error("Failed to get user",
			zap.String("user_id", id),
			zap.Error(err))
		return h.handleError(c, err)
	}

	logger.Info("User retrieved successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))

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
	// Enrich logger with trace context
	logger := h.logger.WithTraceContext(c.UserContext())

	id := c.Params("id")

	// Get validated data from middleware
	req := c.Locals("validated_data").(dto.UpdateUserRequestDto)

	logger.Info("Updating user",
		zap.String("user_id", id),
		zap.String("new_name", *req.Name),
		zap.String("new_email", *req.Email))

	// Call service
	user, err := h.service.UpdateUser(
		c.UserContext(),
		id,
		*req.Name,
		*req.Email,
	)

	if err != nil {
		logger.Error("Failed to update user",
			zap.String("user_id", id),
			zap.Error(err))
		return h.handleError(c, err)
	}
	h.metrics.UsersUpdated.Inc()

	logger.Info("User updated successfully",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))

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
	// Enrich logger with trace context
	logger := h.logger.WithTraceContext(c.UserContext())

	id := c.Params("id")
	logger.Info("Deleting user", zap.String("user_id", id))

	err := h.service.DeleteUser(c.UserContext(), id)
	if err != nil {
		logger.Error("Failed to delete user",
			zap.String("user_id", id),
			zap.Error(err))
		return h.handleError(c, err)
	}
	h.metrics.UsersDeleted.Inc()

	logger.Info("User deleted successfully", zap.String("user_id", id))
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
	// Enrich logger with trace context
	logger := h.logger.WithTraceContext(c.UserContext())

	// Get validated query params from middleware
	query := c.Locals("validated_query").(dto.ListUsersQueryDto)

	logger.Info("Listing users",
		zap.Int("limit", query.Limit),
		zap.Int("offset", query.Offset))

	// Call service
	users, total, err := h.service.ListUsers(c.UserContext(), query.Limit, query.Offset)
	if err != nil {
		logger.Error("Failed to list users", zap.Error(err))
		return h.handleError(c, err)
	}

	logger.Info("Users listed successfully",
		zap.Int64("total", total),
		zap.Int("returned", len(users)))

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
