package middleware

import (
	"github.com/PPLGPride/Be-Ambis-Solving/internal/authz"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func BoardAccessByBoardPath(param string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := utils.UserIDFromCtx(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		bid, err := primitive.ObjectIDFromHex(c.Params(param))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid board id"})
		}
		ctx, cancel := authz.WithTimeout(c.Context())
		defer cancel()
		ok, e := authz.IsMemberOrOwner(ctx, bid, uid)
		if e != nil {
			return c.Status(500).JSON(fiber.Map{"error": e.Error()})
		}
		if !ok {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}

func BoardAccessByTaskPath(taskParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := utils.UserIDFromCtx(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		tid, err := primitive.ObjectIDFromHex(c.Params(taskParam))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid task id"})
		}
		ctx, cancel := authz.WithTimeout(c.Context())
		defer cancel()
		bid, e := authz.BoardIDFromTask(ctx, tid)
		if e != nil {
			return c.Status(404).JSON(fiber.Map{"error": "task not found"})
		}
		ok, e := authz.IsMemberOrOwner(ctx, bid, uid)
		if e != nil {
			return c.Status(500).JSON(fiber.Map{"error": e.Error()})
		}
		if !ok {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}

// Guard akses board lewat QUERY ?boardId=...
func BoardAccessByBoardQuery(queryKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		bidStr := c.Query(queryKey)
		if bidStr == "" {
			// tidak ada filter boardId → lewati guard
			return c.Next()
		}
		uid, err := utils.UserIDFromCtx(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		bid, err := primitive.ObjectIDFromHex(bidStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid board id"})
		}
		ctx, cancel := authz.WithTimeout(c.Context())
		defer cancel()
		ok, e := authz.IsMemberOrOwner(ctx, bid, uid)
		if e != nil {
			return c.Status(500).JSON(fiber.Map{"error": e.Error()})
		}
		if !ok {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}
