package handlers

import (
	"context"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type DevHandler struct {
	Boards services.BoardService
	Tasks  services.TaskService
}

func NewDevHandler(b services.BoardService, t services.TaskService) *DevHandler {
	return &DevHandler{Boards: b, Tasks: t}
}

func (h *DevHandler) Seed(c *fiber.Ctx) error {
	uid, err := utils.UserIDFromCtx(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// 1) Buat board + default columns
	desc := "Demo board for FE integration"
	b, err := h.Boards.Create(ctx, uid, "Project Demo", &desc, nil, nil)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// ambil kolom pertama
	if len(b.Columns) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "board has no columns"})
	}
	col0 := b.Columns[0].ID

	// 2) Beberapa task awal
	_, _ = h.Tasks.Create(ctx, b.ID, uid, "Setup API", nil, col0, nil, nil, nil)
	_, _ = h.Tasks.Create(ctx, b.ID, uid, "Wire Frontend", nil, col0, nil, nil, nil)
	_, _ = h.Tasks.Create(ctx, b.ID, uid, "Write README", nil, col0, nil, nil, nil)

	return c.JSON(fiber.Map{"board": b})
}
