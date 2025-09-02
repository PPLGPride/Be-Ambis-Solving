// Be-Ambis-Solving/internal/handlers/task_handler.go

package handlers

import (
	"context"
	"log" // Impor log untuk debugging
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/models"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
	socketio "github.com/googollee/go-socket.io" // Impor socket.io
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 1. Tambahkan SocketServer ke dalam struct
type TaskHandler struct {
	Svc          services.TaskService
	SocketServer *socketio.Server
}

// 2. Modifikasi constructor untuk menerima SocketServer
func NewTaskHandler(s services.TaskService, so *socketio.Server) *TaskHandler {
	return &TaskHandler{Svc: s, SocketServer: so}
}

// (Tipe taskCreateReq tidak berubah)
type taskCreateReq struct {
	// ...
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	ColumnID    string             `json:"columnId"`
	Status      *models.TaskStatus `json:"status"`
	DueDate     *time.Time         `json:"dueDate"`
	Assignees   []string           `json:"assignees"`
}

func (h *TaskHandler) Create(c *fiber.Ctx) error {
	boardID, err := utils.MustObjectID(c.Params("boardId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid boardId"})
	}
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req taskCreateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	var assignees []primitive.ObjectID
	for _, a := range req.Assignees {
		if oid, err := primitive.ObjectIDFromHex(a); err == nil {
			assignees = append(assignees, oid)
		}
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	t, err := h.Svc.Create(ctx, boardID, uid, req.Title, req.Description, req.ColumnID, req.Status, req.DueDate, assignees)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. Broadcast event setelah berhasil
	h.SocketServer.BroadcastToNamespace("/", "board_updated", nil)
	log.Println("Broadcast [board_updated] setelah Create Task")

	return c.Status(201).JSON(t)
}

// (Handler ListByBoard dan Get tidak perlu broadcast)
func (h *TaskHandler) ListByBoard(c *fiber.Ctx) error {
	// ... kode tidak berubah ...
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

func (h *TaskHandler) Get(c *fiber.Ctx) error {
	// ... kode tidak berubah ...
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	t, err := h.Svc.Get(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(t)
}

// (Tipe taskUpdateReq tidak berubah)
type taskUpdateReq struct {
	// ...
	Title       *string              `json:"title"`
	Description *string              `json:"description"`
	Status      *models.TaskStatus   `json:"status"`
	Priority    *models.TaskPriority `json:"priority"`
	ColumnID    *string              `json:"columnId"`
	DueDate     *time.Time           `json:"dueDate"`
	StartDate   *time.Time           `json:"startDate"`
	Tags        *[]string            `json:"tags"`
}

func (h *TaskHandler) Update(c *fiber.Ctx) error {
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req taskUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	patch := bson.M{}
	if req.Title != nil {
		patch["title"] = *req.Title
	}
	if req.Description != nil {
		patch["description"] = req.Description
	}
	if req.Status != nil {
		patch["status"] = *req.Status
	}
	if req.Priority != nil {
		patch["priority"] = *req.Priority
	}
	if req.ColumnID != nil {
		patch["columnId"] = *req.ColumnID
	}
	if req.DueDate != nil {
		patch["dueDate"] = req.DueDate
	}
	if req.StartDate != nil {
		patch["startDate"] = req.StartDate
	}
	if req.Tags != nil {
		patch["tags"] = *req.Tags
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	if err := h.Svc.Update(ctx, id, patch, uid); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. Broadcast event setelah berhasil
	h.SocketServer.BroadcastToNamespace("/", "board_updated", nil)
	log.Println("Broadcast [board_updated] setelah Update Task")

	return c.SendStatus(204)
}

func (h *TaskHandler) Delete(c *fiber.Ctx) error {
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	if err := h.Svc.Delete(ctx, id); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. Broadcast event setelah berhasil
	h.SocketServer.BroadcastToNamespace("/", "board_updated", nil)
	log.Println("Broadcast [board_updated] setelah Delete Task")

	return c.SendStatus(204)
}

// (Tipe moveReq tidak berubah)
type moveReq struct {
	ToColumnID string `json:"toColumnId"`
	ToPosition int    `json:"toPosition"`
}

func (h *TaskHandler) Move(c *fiber.Ctx) error {
	id, err := utils.MustObjectID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	var req moveReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.ToPosition < 1 {
		req.ToPosition = 1
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	if err := h.Svc.Move(ctx, id, req.ToColumnID, req.ToPosition); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. Broadcast event setelah berhasil
	h.SocketServer.BroadcastToNamespace("/", "board_updated", nil)
	log.Println("Broadcast [board_updated] setelah Move Task")

	return c.SendStatus(204)
}
