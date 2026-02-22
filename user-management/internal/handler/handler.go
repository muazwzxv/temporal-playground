package handler

import (
	"github.com/gofiber/fiber/v2"
	healthHandler "github.com/muazwzxv/user-management/internal/handler/health"
	userHandler "github.com/muazwzxv/user-management/internal/handler/user"
	"github.com/samber/do/v2"
)

func InjectHandler(i do.Injector, app *fiber.App) {
	// Provide handlers
	do.Provide(i, healthHandler.NewHealthHandler)
	do.Provide(i, userHandler.NewUserHandler)

	// Invoke handlers from DI container and register their routes
	do.MustInvoke[*healthHandler.HealthHandler](i).RegisterRoutes(app)
	do.MustInvoke[*userHandler.UserHandler](i).RegisterRoutes(app)
}
