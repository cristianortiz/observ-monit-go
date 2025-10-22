package response

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return Error(c, fiber.StatusBadRequest, "bad_request", "Invalid input")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]string
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "bad_request", response["error"])
	assert.Equal(t, "Invalid input", response["message"])
}

func TestValidationError(t *testing.T) {
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		fields := fiber.Map{
			"email":    "must be a valid email",
			"password": "must be at least 8 characters",
		}
		return ValidationError(c, "Validation failed", fields)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "validation_error", response["error"])
	assert.Equal(t, "Validation failed", response["message"])

	fields := response["fields"].(map[string]interface{})
	assert.Equal(t, "must be a valid email", fields["email"])
	assert.Equal(t, "must be at least 8 characters", fields["password"])
}

func TestNotFound(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return NotFound(c, "User not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]string
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "not_found", response["error"])
	assert.Equal(t, "User not found", response["message"])
}

func TestUnauthorized(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return Unauthorized(c, "Invalid credentials")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]string
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "unauthorized", response["error"])
	assert.Equal(t, "Invalid credentials", response["message"])
}

func TestInternalError(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return InternalError(c, "Something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]string
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "internal_error", response["error"])
	assert.Equal(t, "Something went wrong", response["message"])
}
