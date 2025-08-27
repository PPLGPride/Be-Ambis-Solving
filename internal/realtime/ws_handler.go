package realtime

import (
	"github.com/gofiber/websocket/v2"
)

func WSHandler(c *websocket.Conn) {
	room := c.Params("boardId")
	if room == "" {
		_ = c.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","data":"missing boardId"}`))
		c.Close()
		return
	}
	H.Join(room, c)
	defer func() {
		H.Leave(room, c)
		c.Close()
	}()

	// keep connection; ignore client messages
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
