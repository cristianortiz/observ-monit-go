package middleware

import (
	"errors"
	"slices"

	"github.com/cristianortiz/observ-monit-go/pkg/config"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims define la estructura de los datos dentro del JWT
// Auth0 usa "namespaced claims" por eso los campos tienen URLs
type CustomClaims struct {
	jwt.RegisteredClaims          // Claims estándar: iss, sub, aud, exp, iat
	Email                string   `json:"email"`                                // Email del usuario
	Name                 string   `json:"name"`                                 // Nombre del usuario
	TenantID             string   `json:"https://api.factorit.com/tenant_id"`   // ID de la empresa (multi-tenancy)
	TenantName           string   `json:"https://api.factorit.com/tenant_name"` // Nombre de la empresa´
	Roles                []string `json:"https://api.factorit.com/roles"`       // Roles: ["admin", "user"]
	Permissions          []string `json:"https://api.factorit.com/permissions"` // Permisos: ["read:users", "write:users"]
}

// RequireAuth middleware valida JWT tokens desde Auth0
// Si AUTH_ENABLED=false, permite pasar sin autenticación (desarrollo)
// Si AUTH_ENABLED=true, valida firma, expiración, audience e issuer
func RequireAuth(cfg *config.Config) fiber.Handler {
	// En desarrollo podemos desactivar auth completamente
	if !cfg.Security.AuthEnabled {
		return func(c *fiber.Ctx) error {
			// Modo desarrollo: inyectar usuario fake para testing
			c.Locals("user_id", "dev-user-123")
			c.Locals("user_email", "dev@example.com")
			c.Locals("user_name", "Development User")
			return c.Next()
		}
	}

	// Configuración del middleware JWT de Fiber
	return jwtware.New(jwtware.Config{
		// JWKS URL: Auto-descarga las claves públicas de Auth0
		// y las refresca automáticamente cuando rotan
		JWKSetURLs: []string{cfg.Security.AuthJWKSUrl},

		// Claims personalizados (Auth0 incluye email, name, roles, etc)
		Claims: &CustomClaims{},

		// Validar que el token fue emitido PARA nuestra API
		TokenLookup: "header:Authorization", // Buscar en header "Authorization: Bearer <token>"
		AuthScheme:  "Bearer",

		// Error handler: cuando el token es inválido, expirado o falta
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log del error (útil para debugging)
			// logger.Warn("Auth failed", zap.Error(err), zap.String("ip", c.IP()))

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid or expired JWT token",
			})
		},

		// Success handler: cuando el token es válido
		// Aquí extraemos los datos del token y los inyectamos en el context
		SuccessHandler: func(c *fiber.Ctx) error {
			// Obtener el token parseado
			user := c.Locals("user").(*jwt.Token)
			claims := user.Claims.(*CustomClaims)

			// Validar audience (¿el token fue emitido para nuestra API?)
			// Audience es un slice, verificamos si contiene nuestro audience
			if cfg.Security.AuthAudience != "" {
				validAudience := slices.Contains(claims.Audience, cfg.Security.AuthAudience)
				if !validAudience {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":   "forbidden",
						"message": "Invalid token audience",
					})
				}
			}

			// Validar issuer (¿el token vino de nuestro Auth0 tenant?)
			if cfg.Security.AuthIssuer != "" {
				if claims.Issuer != cfg.Security.AuthIssuer {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":   "forbidden",
						"message": "Invalid token issuer",
					})
				}
			}

			// Inyectar datos del usuario en el context
			// Los handlers pueden acceder a estos datos con c.Locals("user_id")
			c.Locals("user_id", claims.Subject) // subject = user ID único
			c.Locals("user_email", claims.Email)
			c.Locals("user_name", claims.Name)
			c.Locals("user_roles", claims.Roles)
			c.Locals("user_permissions", claims.Permissions)

			// Multi-tenancy: inyectar tenant_id si existe
			if claims.TenantID != "" {
				c.Locals("tenant_id", claims.TenantID)
				c.Locals("tenant_name", claims.TenantName)
			}

			//Log de auditoría: quién accedió a qué
			// logger.Info("User authenticated",
			// 	zap.String("user_id", claims.Subject),
			// 	zap.String("email", claims.Email),
			// 	zap.String("method", c.Method()),
			// 	zap.String("path", c.Path()),
			// 	zap.Strings("permissions", claims.Permissions),
			// )

			return c.Next()
		},
	})
}

// RequirePermission valida que el usuario tenga un permiso específico
// Ejemplo: RequirePermission("delete:users") bloquea si no tiene ese scope
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, ok := c.Locals("user_permissions").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "No permissions found in token",
			})
		}

		// Verificar si tiene el permiso requerido
		if slices.Contains(permissions, permission) {
			return c.Next()
		}

		// No tiene permiso
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "forbidden",
			"message": "Insufficient permissions: requires " + permission,
		})
	}
}

// RequireRole valida que el usuario tenga un rol específico
// Ejemplo: RequireRole("admin") solo permite admins
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roles, ok := c.Locals("user_roles").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "No roles found in token",
			})
		}

		// Verificar si tiene el rol requerido
		if slices.Contains(roles, role) {
			return c.Next()
		}

		//  No tiene rol
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "forbidden",
			"message": "Requires role: " + role,
		})
	}
}

// GetUserID extrae el user_id del context (helper para handlers)
func GetUserID(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return "", errors.New("user_id not found in context")
	}
	return userID, nil
}

// GetTenantID extrae el tenant_id del context (multi-tenancy)
func GetTenantID(c *fiber.Ctx) (string, error) {
	tenantID, ok := c.Locals("tenant_id").(string)
	if !ok || tenantID == "" {
		return "", errors.New("tenant_id not found in context")
	}
	return tenantID, nil
}
