package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	service "github.com/muazwzxv/user-management/internal/service/user"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(i do.Injector) (*UserHandler, error) {
	svc := do.MustInvoke[service.UserService](i)

	return &UserHandler{
		service: svc,
	}, nil
}

func (h *UserHandler) RegisterRoutes(app *fiber.App) {

	app.Post("/api/v1/users", h.CreateUser)
}
