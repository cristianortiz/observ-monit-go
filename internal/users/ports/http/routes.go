package http

import (
	"github.com/cristianortiz/observ-monit-go/internal/users/ports/http/dto"
	"github.com/cristianortiz/observ-monit-go/pkg/http-utils/middleware"
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all user routes
func RegisterRoutes(app *fiber.App, handler *UserHandler, basePath string) {
	api := app.Group(basePath)
	users := api.Group("/users")

	// CRUD operations
	users.Post("/",
		middleware.ValidateBody[dto.CreateUserRequestDto](),
		handler.CreateUser,
	)

	users.Get("/",
		middleware.ValidateQuery[dto.ListUsersQueryDto](),
		handler.ListUsers,
	)

	users.Get("/:id",
		middleware.ValidateParam("id", "uuid"),
		handler.GetUser)

	users.Put("/:id",
		middleware.ValidateParam("id", "uuid"),
		middleware.ValidateBody[dto.UpdateUserRequestDto](),
		handler.UpdateUser,
	)

	users.Delete("/:id",
		middleware.ValidateParam("id", "uuid"),
		handler.DeleteUser,
	)
}
