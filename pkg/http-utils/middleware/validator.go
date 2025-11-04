package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var validate *validator.Validate

// InitValidator initializes the global validator instance
//
// Call this once in main.go before starting the server
func InitValidator() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	// Register custom UUID validator
	validate.RegisterValidation("uuid", validateUUID)
}

// ValidateStruct validates any struct using generics
func ValidateStruct[T any](data T) (fiber.Map, error) {
	if err := validate.Struct(data); err != nil {
		fieldErrors := getValidationErrors(err)
		return fieldErrors, fmt.Errorf("validation failed")
	}
	return nil, nil
}

// getValidationErrors converts validator errors to fiber.Map
func getValidationErrors(err error) fiber.Map {
	fields := fiber.Map{}
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			fields[strings.ToLower(e.Field())] = e.Error()
		}
	}
	return fields
}

// ValidateBody is a Fiber middleware that validates request body
//
// Usage in any module:
//
//	import "github.com/cristianortiz/observ-monit-go/pkg/httputil/middleware"
//
//	app.Post("/users", middleware.ValidateBody[dto.CreateUserRequest](), handler)
func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_json",
				"message": "Failed to parse request body",
			})
		}

		if fieldErrors, err := ValidateStruct(data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Request validation failed",
				"fields":  fieldErrors,
			})
		}

		c.Locals("validated_data", data)
		return c.Next()
	}
}

// ValidateQuery is a Fiber middleware that validates query parameters
//
// Usage in any module:
//
//	app.Get("/users", middleware.ValidateQuery[dto.ListUsersQuery](), handler)
func ValidateQuery[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var query T

		if err := c.QueryParser(&query); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_query",
				"message": "Failed to parse query parameters",
			})
		}

		// Set defaults if available
		if setter, ok := any(&query).(interface{ SetDefaults() }); ok {
			setter.SetDefaults()
		}

		if fieldErrors, err := ValidateStruct(query); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Query validation failed",
				"fields":  fieldErrors,
			})
		}

		c.Locals("validated_query", query)
		return c.Next()
	}
}

// validateUUID is a Custom validator to check if a string is a valid UUID
func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, err := uuid.Parse(value)
	return err == nil
}

// ValidateParam validates a specific route parameter
// Usage:
//
//	app.Get("/users/:id", middleware.ValidateParam("id", "uuid"), handler)
func ValidateParam(paramName string, validationType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		value := c.Params(paramName)

		switch validationType {
		case "uuid":
			if _, err := uuid.Parse(value); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "validation_error",
					"message": fmt.Sprintf("Invalid %s format", paramName),
					"fields": fiber.Map{
						paramName: fmt.Sprintf("must be a valid UUID, got: %s", value),
					},
				})
			}
		case "numeric":
			if _, err := strconv.Atoi(value); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "validation_error",
					"message": fmt.Sprintf("Invalid %s format", paramName),
					"fields": fiber.Map{
						paramName: "must be a number",
					},
				})
			}
		default:
			return fmt.Errorf("unknown validation type: %s", validationType)
		}

		return c.Next()
	}
}
