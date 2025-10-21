package response

import "github.com/gofiber/fiber/v2"

// Error standardizes error responses across all modules
func Error(c *fiber.Ctx, status int, errorType, message string, fields ...fiber.Map) error {
	response := fiber.Map{
		"error":   errorType,
		"message": message,
	}

	if len(fields) > 0 && fields[0] != nil {
		response["fields"] = fields[0]
	}

	return c.Status(status).JSON(response)
}

// ValidationError returns a 400 validation error
func ValidationError(c *fiber.Ctx, message string, fields fiber.Map) error {
	return Error(c, fiber.StatusBadRequest, "validation_error", message, fields)
}

// NotFound returns a 404 not found error
func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, "not_found", message)
}

// Unauthorized returns a 401 unauthorized error
func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, "unauthorized", message)
}

// InternalError returns a 500 internal server error
func InternalError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, "internal_error", message)
}
