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

func TestSuccess(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		data := map[string]string{
			"message": "success",
		}
		return Success(c, data)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]string
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "success", response["message"])
}

func TestCreated(t *testing.T) {
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		data := map[string]interface{}{
			"id":   "123",
			"name": "Test",
		}
		return Created(c, data)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "123", response["id"])
	assert.Equal(t, "Test", response["name"])
}

func TestNoContent(t *testing.T) {
	app := fiber.New()

	app.Delete("/test", func(c *fiber.Ctx) error {
		return NoContent(c)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	assert.Empty(t, bodyBytes)
}

func TestMessage(t *testing.T) {
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		return Message(c, "Operation completed successfully")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]string
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "Operation completed successfully", response["message"])
}

func TestSuccess_WithComplexData(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		data := map[string]interface{}{
			"user": map[string]string{
				"id":   "123",
				"name": "John",
			},
			"count":  42,
			"active": true,
		}
		return Success(c, data)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(bodyBytes, &response)

	user := response["user"].(map[string]interface{})
	assert.Equal(t, "123", user["id"])
	assert.Equal(t, "John", user["name"])
	assert.Equal(t, float64(42), response["count"])
	assert.Equal(t, true, response["active"])
}
