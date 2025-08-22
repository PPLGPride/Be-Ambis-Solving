package middleware

import (
	"strings"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// internal/middleware/auth.go
func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if h == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid auth scheme"})
		}
		tokenStr := parts[1]
		claims, err := utils.ParseJWT(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals("userId", claims.Subject)
		return c.Next()
	}
}
