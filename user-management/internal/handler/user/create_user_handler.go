package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/user-management/internal/dto/request"
	"github.com/muazwzxv/user-management/internal/dto/response"
)

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	logger := log.WithContext(c.UserContext())

	var req request.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warnw("invalid request body",
			"error", err,
			"path", c.Path(),
			"ip", c.IP())
		return response.HandleError(c, response.BuildErrorWithCode(
			fiber.StatusBadRequest,
			"Invalid request body",
			"INVALID_REQUEST_BODY",
		))
	}

	logger.Infow("creating user",
		"name", req.Name,
		"referenceID", req.ReferenceID)

	result, err := h.service.CreateUser(c.Context(), req)
	if err != nil {
		logger.Errorw("failed to create user",
			"error", err,
			"name", req.Name)
		return response.HandleError(c, err)
	}

	logger.Infow("user created successfully",
		"id", result.ID,
		"name", result.Name)

	return c.Status(fiber.StatusCreated).JSON(result)
}
