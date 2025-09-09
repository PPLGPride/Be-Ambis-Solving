package httpx

import (
	"github.com/PPLGPride/Be-Ambis-Solving/internal/validation"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type APIError struct {
	Error  string            `json:"error"`
	Code   string            `json:"code,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

func BadRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(APIError{Error: msg, Code: "bad_request"})
}
func Unauthorized(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(APIError{Error: msg, Code: "unauthorized"})
}
func Forbidden(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusForbidden).JSON(APIError{Error: msg, Code: "forbidden"})
}
func NotFound(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusNotFound).JSON(APIError{Error: msg, Code: "not_found"})
}
func ServerError(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(APIError{Error: msg, Code: "server_error"})
}

func FromValidation(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	out := APIError{Error: "validation_failed", Code: "validation", Fields: map[string]string{}}
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verrs {
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				out.Fields[field] = "required"
			case "email":
				out.Fields[field] = "invalid_email"
			case "min":
				out.Fields[field] = "min"
			case "oneof":
				out.Fields[field] = "invalid_value"
			default:
				out.Fields[field] = fe.Tag()
			}
		}
		return c.Status(fiber.StatusBadRequest).JSON(out)
	}
	return BadRequest(c, err.Error())
}

// Shorthand untuk validate struct
func ValidateBody[T any](c *fiber.Ctx, dst *T) error {
	if err := c.BodyParser(dst); err != nil {
		return BadRequest(c, "invalid_body")
	}
	if err := validation.V.Struct(dst); err != nil {
		return FromValidation(c, err)
	}
	return nil
}
