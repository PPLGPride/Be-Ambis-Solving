package handlers

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/httpx"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/models"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
	socketio "github.com/googollee/go-socket.io"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==============================
// Handler & ctor
// ==============================
type TaskHandler struct {
	Svc    services.TaskService
	Socket *socketio.Server
}

func NewTaskHandler(s services.TaskService, sock *socketio.Server) *TaskHandler {
	return &TaskHandler{Svc: s, Socket: sock}
}

// ==============================
// DTOs
// ==============================
type taskCreateReq struct {
	Title       string  `json:"title"`
	ColumnID    string  `json:"columnId"`
	Description *string `json:"description"`
}

type taskMoveReq struct {
	ToColumnID string `json:"toColumnId"`
	ToPosition int    `json:"toPosition"` // 1-based
}

type taskUpdateReq struct {
	Title       *string              `json:"title"`
	Description *string              `json:"description"`
	Status      *models.TaskStatus   `json:"status"`
	Priority    *models.TaskPriority `json:"priority"`
	ColumnID    *string              `json:"columnId"`
	DueDate     *time.Time           `json:"dueDate"`
	StartDate   *time.Time           `json:"startDate"`
	Tags        *[]string            `json:"tags"`
}

// ==============================
// Helpers
// ==============================
func userIDFromCtx(c *fiber.Ctx) (primitive.ObjectID, error) {
	v := c.Locals("userId")
	switch t := v.(type) {
	case primitive.ObjectID:
		return t, nil
	case string:
		return primitive.ObjectIDFromHex(t)
	default:
		return primitive.NilObjectID, fiber.ErrUnauthorized
	}
}

func mustOIDParam(c *fiber.Ctx, key string) (primitive.ObjectID, error) {
	p := strings.TrimSpace(c.Params(key))
	return primitive.ObjectIDFromHex(p)
}

// ==============================
// Handlers
// ==============================

// GET /api/boards/:boardId/tasks
func (h *TaskHandler) ListByBoard(c *fiber.Ctx) error {
	boardID, err := mustOIDParam(c, "boardId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid boardId"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 6*time.Second)
	defer cancel()

	items, err := h.Svc.ListByBoard(ctx, boardID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

// GET /api/tasks/:id
func (h *TaskHandler) Get(c *fiber.Ctx) error {
	tid, err := mustOIDParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 6*time.Second)
	defer cancel()

	t, err := h.Svc.Get(ctx, tid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(t)
}

// POST /api/boards/:boardId/tasks
func (h *TaskHandler) Create(c *fiber.Ctx) error {
	boardID, err := mustOIDParam(c, "boardId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid boardId"})
	}

	uid, err := userIDFromCtx(c) // ← ambil 2 nilai (uid, err)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req taskCreateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	req.Title = strings.TrimSpace(req.Title)
	req.ColumnID = strings.TrimSpace(req.ColumnID)
	if len(req.Title) < 1 || req.ColumnID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title & columnId required"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	// Service Create signature (berdasarkan error yang kamu kirim):
	// want (ctx, boardID, userID, title, *description, columnId, *status, *dueDate, []primitive.ObjectID)
	var status *models.TaskStatus = nil
	var due *time.Time = nil
	var assignees []primitive.ObjectID = nil

	t, err := h.Svc.Create(ctx, boardID, uid, req.Title, req.Description, req.ColumnID, status, due, assignees)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Broadcast
	if h.Socket != nil {
		h.Socket.BroadcastToRoom("/", t.BoardID.Hex(), "task_created", fiber.Map{
			"id":       t.ID.Hex(),
			"boardId":  t.BoardID.Hex(),
			"title":    t.Title,
			"columnId": t.ColumnID,
			"order":    t.Order,
			"actorId":  uid.Hex(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(t)
}

// PATCH /api/tasks/:id
func (h *TaskHandler) Update(c *fiber.Ctx) error {
	tid, err := mustOIDParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	uid, err := userIDFromCtx(c) // ← ambil 2 nilai
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req taskUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	// Bangun update doc sesuai field yg dikirim (pakai bson.M, sesuai signature servicemu)
	update := bson.M{"updatedAt": time.Now()}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		update["title"] = title
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Status != nil {
		update["status"] = *req.Status
	}
	if req.Priority != nil {
		update["priority"] = *req.Priority
	}
	if req.ColumnID != nil {
		col := strings.TrimSpace(*req.ColumnID)
		if col != "" {
			update["columnId"] = col
		}
	}
	if req.DueDate != nil {
		update["dueDate"] = *req.DueDate
	}
	if req.StartDate != nil {
		update["startDate"] = *req.StartDate
	}
	if req.Tags != nil {
		update["tags"] = *req.Tags
	}

	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	// Service Update signature (dari error): want (ctx, id, bson.M, userID)
	if err := h.Svc.Update(ctx, tid, update, uid); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Ambil lagi untuk broadcast
	t, err := h.Svc.Get(ctx, tid)
	if err != nil {
		return c.SendStatus(204)
	}

	if h.Socket != nil {
		h.Socket.BroadcastToRoom("/", t.BoardID.Hex(), "task_updated", fiber.Map{
			"id":      t.ID.Hex(),
			"boardId": t.BoardID.Hex(),
			"actorId": uid.Hex(),
		})
	}
	if t, getErr := h.Svc.Get(ctx, tid); getErr == nil && h.Socket != nil {
		room := t.BoardID.Hex()
		log.Printf("[SOCKET][EMIT] room=%s event=%s id=%s\n", room, "task_updated", t.ID.Hex())
		h.Socket.BroadcastToRoom("/", room, "task_updated", fiber.Map{
			"id":      t.ID.Hex(),
			"actorId": c.Locals("userId"),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DELETE /api/tasks/:id
func (h *TaskHandler) Delete(c *fiber.Ctx) error {
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized")
	}

	tid, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return httpx.BadRequest(c, "invalid id")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	// Ambil dulu task utk tahu board room (jika perlu)
	var room string
	if t, _ := h.Svc.Get(ctx, tid); t != nil {
		room = t.BoardID.Hex()
	}

	if err := h.Svc.Delete(ctx, tid); err != nil {
		return httpx.ServerError(c, err.Error())
	}

	if h.Socket != nil {
		h.Socket.BroadcastToRoom("/", room, "task_deleted", fiber.Map{
			"id":      tid.Hex(),
			"boardId": room,
			"actorId": uid.Hex(),
		})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// POST /api/tasks/:id/move
func (h *TaskHandler) Move(c *fiber.Ctx) error {
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	tid, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	var req struct {
		ToColumnID string `json:"toColumnId" validate:"required"`
		ToPosition int    `json:"toPosition" validate:"required,min=1"`
	}
	if err := httpx.ValidateBody(c, &req); err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.Svc.Move(ctx, tid, req.ToColumnID, req.ToPosition); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// ⬇️ ambil t untuk broadcast/response
	t, err := h.Svc.Get(ctx, tid)
	if err != nil {
		// fallback broadcast minimal tanpa t
		if h.Socket != nil {
			h.Socket.BroadcastToRoom("/", "UNKNOWN_BOARD", "task_moved", fiber.Map{
				"id":         tid.Hex(),
				"toColumnId": req.ToColumnID,
				"toPosition": req.ToPosition,
				"actorId":    uid.Hex(),
			})
		}
		return c.SendStatus(204)
	}

	if h.Socket != nil {
		h.Socket.BroadcastToRoom("/", t.BoardID.Hex(), "task_moved", fiber.Map{
			"id":         t.ID.Hex(),
			"boardId":    t.BoardID.Hex(),
			"toColumnId": req.ToColumnID,
			"toPosition": req.ToPosition,
			"actorId":    uid.Hex(),
		})
	}

	// return c.SendStatus(204)
	return c.Status(200).JSON(t)
}
