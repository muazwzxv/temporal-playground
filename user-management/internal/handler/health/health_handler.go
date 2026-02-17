package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/samber/do/v2"
	"github.com/muazwzxv/user-management/internal/database"
	"github.com/muazwzxv/user-management/internal/dto/response"
)

type DatabaseService interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	dbService DatabaseService
	version   string
}

func NewHealthHandler(i do.Injector) (*HealthHandler, error) {
	db := do.MustInvoke[*database.Database](i)

	return &HealthHandler{
		dbService: db,
		version:   "1.0.0", // TODO: Make configurable via config
	}, nil
}

func (h *HealthHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.HealthCheck)
	app.Get("/health/ready", h.ReadinessCheck)
}

func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	logger := log.WithContext(c.UserContext())
	logger.Debugw("health check requested",
		"ip", c.IP(),
		"user_agent", c.Get("User-Agent"))

	healthResp := response.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   h.version,
		Services:  make(map[string]response.ServiceHealth),
	}

	if h.dbService != nil {
		if err := h.dbService.Ping(c.Context()); err != nil {
			logger.Warnw("health check database ping failed",
				"error", err)
			healthResp.Services["database"] = response.ServiceHealth{
				Status:  "unhealthy",
				Message: "Database connection failed: " + err.Error(),
			}
			healthResp.Status = "degraded"
		} else {
			healthResp.Services["database"] = response.ServiceHealth{
				Status:  "healthy",
				Message: "Connected",
			}
		}
	}

	statusCode := fiber.StatusOK
	if healthResp.Status == "degraded" {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(healthResp)
}

func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	if h.dbService != nil {
		if err := h.dbService.Ping(c.Context()); err != nil {
			return response.HandleError(c, response.BuildErrorWithCode(
				fiber.StatusServiceUnavailable,
				"Service not ready: database unavailable",
				"SERVICE_UNAVAILABLE",
			))
		}
	}

	return c.JSON(fiber.Map{
		"status": "ready",
		"time":   time.Now().Unix(),
	})
}

func (h *HealthHandler) LivenessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "alive",
		"time":   time.Now().Unix(),
	})
}
