package websocket

import (
	"context"
	"kitis-api/internal/watcher"
	"log/slog"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	conn *websocket.Conn
	send chan []byte
}
type WebSocketHub struct {
	watcher    *watcher.Watcher
	clients    map[*WebSocketClient]bool
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
}

func NewWebSocketHub(watcher *watcher.Watcher) *WebSocketHub {
	return &WebSocketHub{
		watcher:    watcher,
		clients:    make(map[*WebSocketClient]bool),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

func (h *WebSocketHub) Run(shutdownCtx context.Context) {
	for {
		select {

		// server closed
		case <-shutdownCtx.Done():
			for client := range h.clients {
				delete(h.clients, client)
				close(client.send)
			}
			slog.Info("websocket hub closed")
			return

		// register client
		case client := <-h.register:
			h.clients[client] = true

		// unregister client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		// send message
		case message := <-h.watcher.Broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (c *WebSocketClient) ReadPump() {
	defer c.conn.Close()
	updateReadDeadline(c.conn)

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		} else {
			updateReadDeadline(c.conn)
			c.conn.WriteMessage(websocket.TextMessage, []byte("{\"ok\":true}"))
		}
	}
}

func (c *WebSocketClient) WritePump() {
	defer c.conn.Close()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}

	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}
