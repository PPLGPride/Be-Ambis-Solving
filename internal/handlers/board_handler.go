package handlers

import (
	"context"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/models"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BoardHandler struct{ Svc services.BoardService }

func NewBoardHandler(s services.BoardService) *BoardHandler { return &BoardHandler{Svc: s} }

type boardCreateReq struct {
	Name        string               `json:"name"`
	Description *string              `json:"description"`
	Columns     []models.BoardColumn `json:"columns"` // opsional; jika kosong pakai default
	Members     []string             `json:"members"` // hex ObjectID
}

func (h *BoardHandler) Create(c *fiber.Ctx) error {
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req boardCreateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	var memberOIDs []primitive.ObjectID
	for _, m := range req.Members {
		if oid, err := primitive.ObjectIDFromHex(m); err == nil {
			memberOIDs = append(memberOIDs, oid)
		}
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	b, err := h.Svc.Create(ctx, uid, req.Name, req.Description, req.Columns, memberOIDs)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(b)
}

func (h *BoardHandler) List(c *fiber.Ctx) error {
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	out, err := h.Svc.ListForUser(ctx, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func (h *BoardHandler) Get(c *fiber.Ctx) error {
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	b, err := h.Svc.Get(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(b)
}

type boardUpdateReq struct {
	Name        *string               `json:"name"`
	Description *string               `json:"description"`
	Columns     *[]models.BoardColumn `json:"columns"`
	Members     *[]string             `json:"members"`
}

func (h *BoardHandler) Update(c *fiber.Ctx) error {
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var req boardUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	var memberOIDs *[]primitive.ObjectID
	if req.Members != nil {
		tmp := make([]primitive.ObjectID, 0, len(*req.Members))
		for _, m := range *req.Members {
			if oid, err := primitive.ObjectIDFromHex(m); err == nil {
				tmp = append(tmp, oid)
			}
		}
		memberOIDs = &tmp
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	if err := h.Svc.Update(ctx, id, req.Name, req.Description, req.Columns, memberOIDs); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

func (h *BoardHandler) Delete(c *fiber.Ctx) error {
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	if err := h.Svc.Delete(ctx, id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}
