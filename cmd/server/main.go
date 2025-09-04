// Be-Ambis-Solving/cmd/server/main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/handlers"
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

	// --- Socket.IO server ---
	SocketServer = socketio.NewServer(nil)
	SocketServer.OnConnect("/", func(s socketio.Conn) error {
		log.Println("socket connected:", s.ID())
		return nil
	})
	SocketServer.OnEvent("/", "join_board", func(s socketio.Conn, boardID string) {
		log.Printf("socket %s join board %s", s.ID(), boardID)
		s.Join(boardID)
		s.Emit("Joined_board", boardID)
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
	// ------------------------

	app := fiber.New(fiber.Config{AppName: "Be-Ambis-Solving"})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Content-Type, Authorization",
	}))

	// Healthcheck
	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	// DB connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := config.ConnectMongo(ctx); err != nil {
		log.Fatal(err)
	}

	// DI (services & handlers)
	userSvc := services.NewUserService()
	authSvc := services.NewAuthService(userSvc)
	authH := handlers.NewAuthHandler(authSvc, userSvc)

	boardSvc := services.NewBoardService()
	taskSvc := services.NewTaskService()
	noteSvc := services.NewNoteService()

	// Handler yang butuh SocketServer
	boardH := handlers.NewBoardHandler(boardSvc, SocketServer)
	taskH := handlers.NewTaskHandler(taskSvc, SocketServer)

	noteH := handlers.NewNoteHandler(noteSvc)
	timelineH := handlers.NewTimelineHandler()

	// ✅ Dev handler untuk seed data
	devH := handlers.NewDevHandler(boardSvc, taskSvc)

	// ✅ Register routes (7 argumen)
	routes.Register(app, authH, boardH, taskH, noteH, timelineH, devH)

	// Socket.IO endpoint (pakai wildcard untuk long-polling/upgrade)
	app.All("/socket.io/*", adaptor.HTTPHandler(SocketServer))

	log.Fatal(app.Listen(":" + config.Cfg.Port))
}
