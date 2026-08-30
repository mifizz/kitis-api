package websocket

import (
	"kitis-api/internal/config"
	"kitis-api/internal/watcher"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	config  *config.Config
	watcher *watcher.Watcher
}

func NewWebSocketHandler(c *config.Config, w *watcher.Watcher) *WebSocketHandler {
	return &WebSocketHandler{config: c, watcher: w}
}

const readDeadline = 30 * time.Second

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func updateReadDeadline(c *websocket.Conn) {
	c.SetReadDeadline(time.Now().Add(readDeadline))
}
