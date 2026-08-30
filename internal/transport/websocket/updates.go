package websocket

import (
	"context"
	"log/slog"
	"net/http"
)

// / debug
func (h *WebSocketHandler) TestNotification(shutdownCtx context.Context, w http.ResponseWriter, r *http.Request) {
	h.watcher.PushMessage(map[string]any{
		"test": map[string]any{
			"message": "hello",
			"data":    123,
			"status":  true,
		},
	})
}

func (h *WebSocketHandler) Updates(shutdownCtx context.Context, hub *WebSocketHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("upgrade to ws failed", "error", err)
		return
	}

	client := &WebSocketClient{
		conn: conn,
		send: make(chan []byte, 256),
	}
	hub.register <- client

	go client.ReadPump()
	go client.WritePump()

	go func() {
		<-shutdownCtx.Done()
		conn.Close()
	}()
}
