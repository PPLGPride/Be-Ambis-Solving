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

	"github.com/gofiber/adaptor/v2" // <-- PASTIKAN INI DIIMPOR
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	socketio "github.com/googollee/go-socket.io"
)

var SocketServer *socketio.Server

func main() {
	config.Load()

	SocketServer = socketio.NewServer(nil)

	SocketServer.OnConnect("/", func(s socketio.Conn) error {
		log.Println("socket terhubung:", s.ID())
		return nil
	})
	SocketServer.OnError("/", func(s socketio.Conn, e error) {
		log.Println("socket error:", e)
	})
	SocketServer.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("socket terputus:", reason)
	})

	go func() {
		if err := SocketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer SocketServer.Close()

	app := fiber.New(fiber.Config{AppName: "Be-Ambis-Solving"})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Content-Type, Authorization",
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := config.ConnectMongo(ctx); err != nil {
		log.Fatal(err)
	}

	// Dependency Injection
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

	// Daftarkan rute API Anda
	routes.Register(app, authH, boardH, taskH, noteH, timelineH)

	// INI ADALAH BARIS KUNCI:
	// Memberitahu Fiber untuk menggunakan handler Socket.IO untuk rute "/socket.io/"
	app.All("/socket.io/", adaptor.HTTPHandler(SocketServer))

	log.Fatal(app.Listen(":" + config.Cfg.Port))
}
