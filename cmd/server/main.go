package main

import (
	"context"
	"log"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/handlers"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/routes"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	config.Load()

	app := fiber.New(fiber.Config{AppName: "Be-Ambis-Solving"})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // ganti ke domain frontend jika sudah fixed
		AllowHeaders: "Content-Type, Authorization",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Be-Ambis-Solving Backend is running 🚀"})
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := config.ConnectMongo(ctx); err != nil {
		log.Fatal(err)
	}

	// DI
	userSvc := services.NewUserService()
	authSvc := services.NewAuthService(userSvc)
	authH := handlers.NewAuthHandler(authSvc, userSvc)

	boardSvc := services.NewBoardService()
	taskSvc := services.NewTaskService()
	noteSvc := services.NewNoteService()

	boardH := handlers.NewBoardHandler(boardSvc)
	taskH := handlers.NewTaskHandler(taskSvc)
	noteH := handlers.NewNoteHandler(noteSvc)
	timelineH := handlers.NewTimelineHandler()
	devH := handlers.NewDevHandler(boardSvc, taskSvc)

	// RoutesS
	routes.Register(app, authH, boardH, taskH, noteH, timelineH, devH)

	log.Fatal(app.Listen(":" + config.Cfg.Port))
}
