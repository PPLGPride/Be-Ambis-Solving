package main

import (
	"context"
	"log"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/handlers"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/realtime"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/routes"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	socketio "github.com/googollee/go-socket.io"
)

var SocketServer *socketio.Server

func main() {
	config.Load()

	SocketServer := realtime.New()
	go func() {
		if err := SocketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %v", err)
		}
	}()
	SocketServer = socketio.NewServer(nil)
	SocketServer.OnConnect("/", func(s socketio.Conn) error {
		log.Println("socket connected:", s.ID())
		return nil
	})
	SocketServer.OnEvent("/", "join_board", func(s socketio.Conn, boardId string) {
		s.Join(boardId)
		log.Println("[SOCKET] join:", s.ID(), "->", boardId)
	})

	SocketServer.OnEvent("/", "leave_board", func(s socketio.Conn, boardID string) {
		log.Printf("socket %s leave board %s", s.ID(), boardID)
		s.Leave(boardID)
		s.Emit("left_board", boardID)
	})
	SocketServer.OnError("/", func(s socketio.Conn, e error) {
		log.Println("socket error:", e)
	})
	SocketServer.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("socket disconnected:", reason)
	})
	go func() {
		if err := SocketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer SocketServer.Close()

	app := fiber.New(fiber.Config{AppName: "Be-Ambis-Solving"})
	realtime.Mount(app, SocketServer)
	app.All("/socket.io/*", adaptor.HTTPHandler(SocketServer))
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: false,
	}))

	// Healthcheck
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	// DB connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := config.ConnectMongo(ctx); err != nil {
		log.Fatal(err)
	}

	userSvc := services.NewUserService()
	authSvc := services.NewAuthService(userSvc)
	authH := handlers.NewAuthHandler(authSvc, userSvc)

	boardSvc := services.NewBoardService()
	taskSvc := services.NewTaskService()
	noteSvc := services.NewNoteService()

	boardH := handlers.NewBoardHandler(boardSvc, SocketServer)
	taskH := handlers.NewTaskHandler(taskSvc, SocketServer)

	noteH := handlers.NewNoteHandler(noteSvc)
	timelineH := handlers.NewTimelineHandler()

	devH := handlers.NewDevHandler(boardSvc, taskSvc)

	routes.Register(app, authH, boardH, taskH, noteH, timelineH, devH)

	app.Use("/socket.io/*", func(c *fiber.Ctx) error {
		log.Printf("[SOCKETIO] HIT %s", c.OriginalURL())
		return c.Next()
	})

	app.All("/socket.io/*", adaptor.HTTPHandler(SocketServer))

	log.Fatal(app.Listen(":" + config.Cfg.Port))
}
