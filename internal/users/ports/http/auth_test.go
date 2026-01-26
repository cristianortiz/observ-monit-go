package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestAppWithAuth configura Fiber app con autenticación para testing
// NOTA: Este es un setup simplificado. En producción necesitarías:
// 1. Mock del JWKS endpoint de Auth0
// 2. Base de datos de testing
// 3. Configuración completa de dependencies
func setupTestAppWithAuth(t *testing.T) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   "error",
				"message": err.Error(),
			})
		},
	})

	// TODO: Implementar setup completo con:
	// - Mock de Auth0 JWKS
	// - Base de datos de testing
	// - Repositories, services, handlers
	// - Middleware de autenticación

	t.Skip("Test setup pendiente: requiere mock de Auth0 JWKS y DB de testing")
	return app
}

// generateTestToken genera JWT de prueba (solo para testing)
// NOTA: En testing real con Auth0, deberías usar tokens reales o mock del JWKS
func generateTestToken(t *testing.T, claims map[string]interface{}) string {
	// Claims por defecto
	defaultClaims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
		"iss": "https://test.auth0.com/",
		"aud": "https://api.factorit.com",
	}

	// Merge con claims personalizados
	for k, v := range claims {
		defaultClaims[k] = v
	}

	// Firmar con secret de testing (NO usar en producción)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, defaultClaims)
	tokenString, err := token.SignedString([]byte("test-secret-key-for-unit-tests"))
	require.NoError(t, err)

	return tokenString
}

// TestListUsers_Unauthorized verifica que sin token retorna 401
func TestListUsers_Unauthorized(t *testing.T) {
	app := setupTestAppWithAuth(t)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	// No Authorization header

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode, "Should return 401 without token")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "unauthorized", response["error"])
	assert.Contains(t, response["message"], "Invalid or expired JWT token")
}

// TestListUsers_ValidToken verifica que con token válido retorna 200
func TestListUsers_ValidToken(t *testing.T) {
	app := setupTestAppWithAuth(t)

	token := generateTestToken(t, map[string]interface{}{
		"sub":   "test-user-123",
		"email": "test@example.com",
		"scope": "read:users",
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1) // -1 = no timeout
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode, "Should return 200 with valid token")
}

// TestCreateUser_Unauthorized verifica que crear usuario sin token retorna 401
func TestCreateUser_Unauthorized(t *testing.T) {
	app := setupTestAppWithAuth(t)

	payload := `{"email":"test@test.com","name":"Test User","password":"pass123"}`
	req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode)
}

// TestCreateUser_ValidToken verifica que con token válido se puede crear usuario
func TestCreateUser_ValidToken(t *testing.T) {
	app := setupTestAppWithAuth(t)

	token := generateTestToken(t, map[string]interface{}{
		"sub":   "test-user-123",
		"email": "test@example.com",
		"scope": "write:users",
	})

	payload := `{"email":"newuser@test.com","name":"New User","password":"SecurePass123!"}`
	req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, 201, resp.StatusCode, "Should create user with valid token")
}

// TestDeleteUser_ForbiddenWithoutPermission verifica que DELETE sin scope delete:users retorna 403
func TestDeleteUser_ForbiddenWithoutPermission(t *testing.T) {
	app := setupTestAppWithAuth(t)

	// Token válido pero SIN scope delete:users
	token := generateTestToken(t, map[string]interface{}{
		"sub":   "test-user-123",
		"email": "test@example.com",
		"scope": "read:users write:users", // No tiene delete:users
	})

	req := httptest.NewRequest("DELETE", "/api/v1/users/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, 403, resp.StatusCode, "Should return 403 without delete:users permission")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "forbidden", response["error"])
	assert.Contains(t, response["message"], "delete:users")
}

// TestDeleteUser_AllowedWithPermission verifica que DELETE con scope delete:users funciona
func TestDeleteUser_AllowedWithPermission(t *testing.T) {
	app := setupTestAppWithAuth(t)

	// Token con scope delete:users
	token := generateTestToken(t, map[string]interface{}{
		"sub":   "admin-user-123",
		"email": "admin@example.com",
		"scope": "read:users write:users delete:users",
	})

	// Primero crear un usuario para eliminar
	createPayload := `{"email":"todelete@test.com","name":"To Delete","password":"Pass123!"}`
	createReq := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createResp, err := app.Test(createReq, -1)
	require.NoError(t, err)
	require.Equal(t, 201, createResp.StatusCode)

	// Obtener ID del usuario creado
	var createdUser map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createdUser)
	require.NoError(t, err)
	userID := createdUser["id"].(string)

	// Ahora eliminar el usuario
	deleteReq := httptest.NewRequest("DELETE", "/api/v1/users/"+userID, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)

	deleteResp, err := app.Test(deleteReq, -1)
	require.NoError(t, err)

	assert.Equal(t, 204, deleteResp.StatusCode, "Should delete user with delete:users permission")
}

// TestExpiredToken verifica que token expirado retorna 401
func TestExpiredToken(t *testing.T) {
	app := setupTestAppWithAuth(t)

	// Token expirado (exp en el pasado)
	token := generateTestToken(t, map[string]interface{}{
		"sub":   "test-user-123",
		"email": "test@example.com",
		"scope": "read:users",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(), // Expirado hace 1 hora
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode, "Should return 401 with expired token")
}

// TestInvalidToken verifica que token malformado retorna 401
func TestInvalidToken(t *testing.T) {
	app := setupTestAppWithAuth(t)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode, "Should return 401 with invalid token")
}
