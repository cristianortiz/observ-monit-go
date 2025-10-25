package http

import (
	"github.com/cristianortiz/observ-monit-go/internal/users/ports/http/dto"
	"github.com/cristianortiz/observ-monit-go/pkg/http-utils/middleware"
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all user routes
func RegisterRoutes(app *fiber.App, handler *UserHandler) {
	api := app.Group("/api")
	users := api.Group("/users")

	// CRUD operations
	users.Get("/",
		middleware.ValidateQuery[dto.ListUsersQueryDto](), // ← Middleware configurado
		handler.ListUsers,
	)
	users.Post("/", handler.CreateUser)      // POST   /api/users
	users.Get("/:id", handler.GetUser)       // GET    /api/users/:id
	users.Put("/:id", handler.UpdateUser)    // PUT    /api/users/:id
	users.Delete("/:id", handler.DeleteUser) // DELETE /api/users/:id
}
