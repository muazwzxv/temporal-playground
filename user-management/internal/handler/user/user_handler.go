package handler

import (
	"github.com/gofiber/fiber/v2"
	service "github.com/muazwzxv/user-management/internal/service/user"
	"github.com/samber/do/v2"
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

	app.Post("/api/v1/CreateUser", h.CreateUser)
}
