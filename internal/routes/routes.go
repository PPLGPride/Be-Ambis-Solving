package routes

import (
	"github.com/PPLGPride/Be-Ambis-Solving/internal/handlers"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func Register(
	app *fiber.App,
	auth *handlers.AuthHandler,
	boards *handlers.BoardHandler,
	tasks *handlers.TaskHandler,
	notes *handlers.NoteHandler,
	timeline *handlers.TimelineHandler,
	dev *handlers.DevHandler,
) {
	api := app.Group("/api")

	// Public
	api.Post("/login", auth.Login)
	api.Post("/register", auth.Register)

	// Protected (JWT)
	prot := api.Group("", middleware.JWTProtected())

	// Dev seed
	prot.Post("/dev/seed", dev.Seed)

	// Boards
	prot.Post("/boards", boards.Create)
	prot.Get("/boards", boards.List) // list milik user; tak perlu guard tambahan
	prot.Get("/boards/:id", middleware.BoardAccessByBoardPath("id"), boards.Get)
	prot.Patch("/boards/:id", middleware.BoardAccessByBoardPath("id"), boards.Update)
	prot.Delete("/boards/:id", middleware.BoardAccessByBoardPath("id"), boards.Delete)

	// Tasks (scoped by board)
	prot.Get("/boards/:boardId/tasks", middleware.BoardAccessByBoardPath("boardId"), tasks.ListByBoard)
	prot.Post("/boards/:boardId/tasks", middleware.BoardAccessByBoardPath("boardId"), tasks.Create)

	// Single task ops (guard by task -> resolve board)
	prot.Get("/tasks/:id", middleware.BoardAccessByTaskPath("id"), tasks.Get)
	prot.Patch("/tasks/:id", middleware.BoardAccessByTaskPath("id"), tasks.Update)
	prot.Delete("/tasks/:id", middleware.BoardAccessByTaskPath("id"), tasks.Delete)
	prot.Post("/tasks/:id/move", middleware.BoardAccessByTaskPath("id"), tasks.Move)

	// Notes
	prot.Post("/notes", notes.Create) // create boleh; validasi akses dilakukan saat baca
	prot.Get("/boards/:boardId/notes", middleware.BoardAccessByBoardPath("boardId"), notes.ListByBoard)
	prot.Get("/tasks/:taskId/notes", middleware.BoardAccessByTaskPath("taskId"), notes.ListByTask)
	prot.Patch("/notes/:id", notes.Update)
	prot.Delete("/notes/:id", notes.Delete)

	// Timeline (guard jika ada boardId query)
	// Timeline (jika ada ?boardId=, guard member/owner)
	prot.Get("/timeline", middleware.BoardAccessByBoardQuery("boardId"), timeline.Get)

	// Whoami
	prot.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"userId": c.Locals("userId")})
	})
}
