package http

import (
	"github.com/cristianortiz/observ-monit-go/internal/users/ports/http/dto"
	"github.com/cristianortiz/observ-monit-go/pkg/config"
	"github.com/cristianortiz/observ-monit-go/pkg/http-utils/middleware"
	"github.com/cristianortiz/observ-monit-go/pkg/observability/metrics"
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers all user routes with authentication
func RegisterRoutes(app *fiber.App, handler *UserHandler, cfg *config.Config, otelMetrics *metrics.OTELMetrics, basePath string) {
	api := app.Group(basePath)
	users := api.Group("/users")

	//  Aplicar autenticación a TODAS las rutas de usuarios
	users.Use(middleware.RequireAuth(cfg, otelMetrics))

	// CRUD operations (todas protegidas por JWT)
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

	// DELETE requiere permiso específico (solo usuarios con scope delete:users)
	users.Delete("/:id",
		middleware.RequirePermission("delete:users", otelMetrics),
		middleware.ValidateParam("id", "uuid"),
		handler.DeleteUser,
	)
}
