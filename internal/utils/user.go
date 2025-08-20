package utils

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UserIDFromCtx(c *fiber.Ctx) (primitive.ObjectID, error) {
	v := c.Locals("userId")
	s, _ := v.(string)
	if s == "" {
		return primitive.NilObjectID, errors.New("no user in ctx")
	}
	return MustObjectID(s)
}
