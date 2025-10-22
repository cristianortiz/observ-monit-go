package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	InitValidator()
}

// TestDTO for generic testing
type TestCreateRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
}

type TestQueryRequest struct {
	Page     int    `query:"page" validate:"min=1"`
	PageSize int    `query:"page_size" validate:"min=1,max=100"`
	SortBy   string `query:"sort_by" validate:"omitempty,oneof=name email created_at"`
}

func TestValidateStruct_ValidData(t *testing.T) {
	req := TestCreateRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	fieldErrors, err := ValidateStruct(req)

	assert.NoError(t, err)
	assert.Nil(t, fieldErrors)
}

func TestValidateStruct_InvalidData(t *testing.T) {
	tests := []struct {
		name          string
		request       TestCreateRequest
		expectedError bool
		expectedField string
	}{
		{
			name: "missing name",
			request: TestCreateRequest{
				Name:  "",
				Email: "john@example.com",
			},
			expectedError: true,
			expectedField: "name",
		},
		{
			name: "invalid email",
			request: TestCreateRequest{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			expectedError: true,
			expectedField: "email",
		},
		{
			name: "name too short",
			request: TestCreateRequest{
				Name:  "J",
				Email: "john@example.com",
			},
			expectedError: true,
			expectedField: "name",
		},
		{
			name: "name too long",
			request: TestCreateRequest{
				Name:  "This is a very long name that exceeds the maximum allowed length for validation",
				Email: "john@example.com",
			},
			expectedError: true,
			expectedField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldErrors, err := ValidateStruct(tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.NotNil(t, fieldErrors)
				assert.Contains(t, fieldErrors, tt.expectedField)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, fieldErrors)
			}
		})
	}
}

func TestValidateStruct_MultipleErrors(t *testing.T) {
	req := TestCreateRequest{
		Name:  "",              // Invalid
		Email: "invalid-email", // Invalid
	}

	fieldErrors, err := ValidateStruct(req)

	assert.Error(t, err)
	assert.NotNil(t, fieldErrors)
	assert.Contains(t, fieldErrors, "name")
	assert.Contains(t, fieldErrors, "email")
	assert.Len(t, fieldErrors, 2)
}

func TestValidateBodyMiddleware_ValidJSON(t *testing.T) {
	app := fiber.New()

	// Route with middleware
	app.Post("/test", ValidateBody[TestCreateRequest](), func(c *fiber.Ctx) error {
		// check data in context
		data := c.Locals("validated_data")
		require.NotNil(t, data)

		req, ok := data.(TestCreateRequest)
		require.True(t, ok)
		assert.Equal(t, "John Doe", req.Name)
		assert.Equal(t, "john@example.com", req.Email)

		return c.SendStatus(fiber.StatusOK)
	})

	// valid request
	body := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestValidateBodyMiddleware_InvalidJSON(t *testing.T) {
	app := fiber.New()

	app.Post("/test", ValidateBody[TestCreateRequest](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// JSON invalid
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	// check response
	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "invalid_json", response["error"])
	assert.Contains(t, response["message"], "Failed to parse")
}

func TestValidateBodyMiddleware_ValidationFails(t *testing.T) {
	app := fiber.New()

	app.Post("/test", ValidateBody[TestCreateRequest](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// invalid data
	body := map[string]string{
		"name":  "",              // Invalid
		"email": "invalid-email", // Invalid
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "validation_error", response["error"])
	assert.NotNil(t, response["fields"])

	fields := response["fields"].(map[string]interface{})
	assert.Contains(t, fields, "name")
	assert.Contains(t, fields, "email")
}

func TestValidateBodyMiddleware_StopsOnValidationError(t *testing.T) {
	app := fiber.New()

	handlerCalled := false

	app.Post("/test", ValidateBody[TestCreateRequest](), func(c *fiber.Ctx) error {
		handlerCalled = true
		return c.SendStatus(fiber.StatusOK)
	})

	// invalid data
	body := map[string]string{
		"name":  "",
		"email": "invalid",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.False(t, handlerCalled, "Handler should NOT be called on validation error")
}

// ==============================================
// TESTS: ValidateQuery Middleware
// ==============================================

func TestValidateQueryMiddleware_ValidQuery(t *testing.T) {
	app := fiber.New()

	app.Get("/test", ValidateQuery[TestQueryRequest](), func(c *fiber.Ctx) error {
		data := c.Locals("validated_query")
		require.NotNil(t, data)

		query, ok := data.(TestQueryRequest)
		require.True(t, ok)
		assert.Equal(t, 1, query.Page)
		assert.Equal(t, 20, query.PageSize)

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test?page=1&page_size=20", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestValidateQueryMiddleware_InvalidQuery(t *testing.T) {
	app := fiber.New()

	app.Get("/test", ValidateQuery[TestQueryRequest](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// page_size > 100 (invalid)
	req := httptest.NewRequest("GET", "/test?page=1&page_size=200", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]any
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, "validation_error", response["error"])
	fields := response["fields"].(map[string]any)
	assert.Contains(t, fields, "pagesize") // validator lowercase field names
}

func TestValidateQueryMiddleware_InvalidSortBy(t *testing.T) {
	app := fiber.New()

	app.Get("/test", ValidateQuery[TestQueryRequest](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// sort_by not in oneof
	req := httptest.NewRequest("GET", "/test?page=1&page_size=20&sort_by=invalid_field", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ==============================================
// BENCHMARKS
// ==============================================

func BenchmarkValidateStruct(b *testing.B) {
	req := TestCreateRequest{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateStruct(req)
	}
}

func BenchmarkValidateBodyMiddleware(b *testing.B) {
	app := fiber.New()

	app.Post("/test", ValidateBody[TestCreateRequest](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	body := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	bodyJSON, _ := json.Marshal(body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req)
	}
}
