package routes

import (
	"github.com/PPLGPride/Be-Ambis-Solving/internal/handlers"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, auth *handlers.AuthHandler, boards *handlers.BoardHandler, tasks *handlers.TaskHandler) {
	api := app.Group("/api")

	// Public
	api.Post("/login", auth.Login)
	api.Post("/register", auth.Register)

	// Protected
	protected := api.Group("", middleware.JWTProtected())

	// Boards
	protected.Post("/boards", boards.Create)
	protected.Get("/boards", boards.List)
	protected.Get("/boards/:id", boards.Get)
	protected.Patch("/boards/:id", boards.Update)
	protected.Delete("/boards/:id", boards.Delete)

	// Tasks (scoped by board)
	protected.Get("/boards/:boardId/tasks", tasks.ListByBoard)
	protected.Post("/boards/:boardId/tasks", tasks.Create)

	// Single task ops
	protected.Get("/tasks/:id", tasks.Get)
	protected.Patch("/tasks/:id", tasks.Update)
	protected.Delete("/tasks/:id", tasks.Delete)

	// DnD move
	protected.Post("/tasks/:id/move", tasks.Move)

	// Example whoami
	protected.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"userId": c.Locals("userId")})
	})
}
