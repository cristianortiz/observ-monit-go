package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers all user routes
func RegisterRoutes(app *fiber.App, handler *UserHandler) {
	api := app.Group("/api")
	users := api.Group("/users")

	// CRUD operations
	users.Post("/", handler.CreateUser)      // POST   /api/users
	users.Get("/", handler.ListUsers)        // GET    /api/users?limit=20&offset=0
	users.Get("/:id", handler.GetUser)       // GET    /api/users/:id
	users.Put("/:id", handler.UpdateUser)    // PUT    /api/users/:id
	users.Delete("/:id", handler.DeleteUser) // DELETE /api/users/:id
}
