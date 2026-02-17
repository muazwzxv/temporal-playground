package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	HttpCode int    `json:"http_code"`
	Message  string `json:"message"`
	Code     string `json:"code,omitempty"`
}

func BuildError(httpCode int, message string) error {
	return &ErrorResponse{
		HttpCode: httpCode,
		Message:  message,
	}
}

func BuildErrorWithCode(httpCode int, message, code string) error {
	return &ErrorResponse{
		HttpCode: httpCode,
		Message:  message,
		Code:     code,
	}
}

func HandleError(ctx *fiber.Ctx, err error) error {
	var e *ErrorResponse
	if errors.As(err, &e) {
		return ctx.Status(e.HttpCode).JSON(e)
	}

	return ctx.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{
		HttpCode: fiber.StatusInternalServerError,
		Message:  err.Error(),
		Code:     "INTERNAL_ERROR",
	})
}

func (e *ErrorResponse) Error() string {
	return e.Message
}
