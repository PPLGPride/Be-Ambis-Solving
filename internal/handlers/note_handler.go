package handlers

import (
	"context"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NoteHandler struct{ Svc services.NoteService }

func NewNoteHandler(s services.NoteService) *NoteHandler { return &NoteHandler{Svc: s} }

type noteCreateReq struct {
	Content      string     `json:"content"`
	BoardID      *string    `json:"boardId"`      // optional
	TaskID       *string    `json:"taskId"`       // optional
	OnTimelineAt *time.Time `json:"onTimelineAt"` // optional
	Pinned       *bool      `json:"pinned"`       // optional
}

func (h *NoteHandler) Create(c *fiber.Ctx) error {
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req noteCreateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	var bID *primitive.ObjectID
	var tID *primitive.ObjectID
	if req.BoardID != nil && *req.BoardID != "" {
		if oid, err := primitive.ObjectIDFromHex(*req.BoardID); err == nil {
			bID = &oid
		}
	}
	if req.TaskID != nil && *req.TaskID != "" {
		if oid, err := primitive.ObjectIDFromHex(*req.TaskID); err == nil {
			tID = &oid
		}
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	n, err := h.Svc.Create(ctx, uid, req.Content, bID, tID, req.OnTimelineAt, getBool(req.Pinned))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(n)
}

func getBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func (h *NoteHandler) ListByBoard(c *fiber.Ctx) error {
	boardID, err := utils.MustObjectID(c.Params("boardId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid boardId"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	out, err := h.Svc.ListByBoard(ctx, boardID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func (h *NoteHandler) ListByTask(c *fiber.Ctx) error {
	taskID, err := utils.MustObjectID(c.Params("taskId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid taskId"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	out, err := h.Svc.ListByTask(ctx, taskID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

type noteUpdateReq struct {
	Content      *string    `json:"content"`
	OnTimelineAt *time.Time `json:"onTimelineAt"`
	Pinned       *bool      `json:"pinned"`
}

func (h *NoteHandler) Update(c *fiber.Ctx) error {
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var req noteUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	patch := bson.M{}
	if req.Content != nil {
		patch["content"] = *req.Content
	}
	if req.OnTimelineAt != nil {
		patch["onTimelineAt"] = req.OnTimelineAt
	}
	if req.Pinned != nil {
		patch["pinned"] = *req.Pinned
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	if err := h.Svc.Update(ctx, id, patch); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

func (h *NoteHandler) Delete(c *fiber.Ctx) error {
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
