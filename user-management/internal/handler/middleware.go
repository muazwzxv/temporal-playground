package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/muazwzxv/user-management/internal/dto/response"
)

func RequestLoggerMiddleware() fiber.Handler {
	return logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	})
}

func ErrorHandlerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			// Context-aware logging with request details
			logger := log.WithContext(c.UserContext())
			logger.Errorw("request error",
				"method", c.Method(),
				"path", c.Path(),
				"ip", c.IP(),
				"user_agent", c.Get("User-Agent"),
				"error", err)
			return response.HandleError(c, err)
		}
		return nil
	}
}

func CORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

func ContentTypeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() != "/health" && c.Path() != "/ready" && c.Path() != "/live" {
			c.Set("Content-Type", "application/json")
		}
		return c.Next()
	}
}
