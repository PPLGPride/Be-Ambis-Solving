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

	// Protected
	prot := api.Group("", middleware.JWTProtected())
	prot.Post("/dev/seed", dev.Seed)

	// Boards
	prot.Post("/boards", boards.Create)
	prot.Get("/boards", boards.List)
	prot.Get("/boards/:id", boards.Get)
	prot.Patch("/boards/:id", boards.Update)
	prot.Delete("/boards/:id", boards.Delete)

	// Tasks
	prot.Get("/boards/:boardId/tasks", tasks.ListByBoard)
	prot.Post("/boards/:boardId/tasks", tasks.Create)
	prot.Get("/tasks/:id", tasks.Get)
	prot.Patch("/tasks/:id", tasks.Update)
	prot.Delete("/tasks/:id", tasks.Delete)
	prot.Post("/tasks/:id/move", tasks.Move)

	// Notes
	prot.Post("/notes", notes.Create)
	prot.Get("/boards/:boardId/notes", notes.ListByBoard)
	prot.Get("/tasks/:taskId/notes", notes.ListByTask)
	prot.Patch("/notes/:id", notes.Update)
	prot.Delete("/notes/:id", notes.Delete)

	// Timeline
	prot.Get("/timeline", timeline.Get)

	// Whoami
	prot.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"userId": c.Locals("userId")})
	})
}
