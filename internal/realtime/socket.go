// internal/realtime/socket.go
package realtime

import (
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	engineio "github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

func New() *socketio.Server {
	srv := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{}, // fallback awal (handshake via HTTP)
			&websocket.Transport{
				// izinkan semua origin saat dev; perketat di produksi
				CheckOrigin: func(r *http.Request) bool { return true },
			},
		},
	})

	srv.OnConnect("/", func(c socketio.Conn) error {
		log.Printf("[SOCKET] connect id=%s", c.ID())
		return nil
	})
	srv.OnError("/", func(c socketio.Conn, err error) {
		log.Printf("[SOCKET] error id=%s err=%v", c.ID(), err)
	})
	srv.OnDisconnect("/", func(c socketio.Conn, reason string) {
		log.Printf("[SOCKET] disconnect id=%s reason=%s", c.ID(), reason)
	})

	// join/leave room board
	srv.OnEvent("/", "join_board", func(c socketio.Conn, boardID string) {
		c.Join(boardID)
		log.Printf("[SOCKET] join %s -> room=%s", c.ID(), boardID)
	})
	srv.OnEvent("/", "leave_board", func(c socketio.Conn, boardID string) {
		c.Leave(boardID)
		log.Printf("[SOCKET] leave %s -> room=%s", c.ID(), boardID)
	})

	return srv
}

func Mount(app *fiber.App, srv *socketio.Server) {
	// lihat semua hit ke jalur engine.io
	app.Use("/socket.io/*", func(c *fiber.Ctx) error {
		log.Printf("[SOCKETIO] HIT %s", c.OriginalURL())
		return c.Next()
	})
	// W A J I B wildcard
	app.All("/socket.io/*", adaptor.HTTPHandler(srv))
}
