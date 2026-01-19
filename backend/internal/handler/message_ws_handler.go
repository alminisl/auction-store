package handler

import (
	"log"
	"net/http"

	"github.com/auction-cards/backend/internal/middleware"
	ws "github.com/auction-cards/backend/internal/websocket"
	"github.com/google/uuid"
)

type MessageWebSocketHandler struct {
	hub *ws.MessageHub
}

func NewMessageWebSocketHandler(hub *ws.MessageHub) *MessageWebSocketHandler {
	return &MessageWebSocketHandler{hub: hub}
}

// HandleMessageWS handles WebSocket connections for messaging
// Requires authentication - only logged in users can connect
func (h *MessageWebSocketHandler) HandleMessageWS(w http.ResponseWriter, r *http.Request) {
	// Get user ID - this endpoint requires authentication
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := ws.NewMessageClient(h.hub, conn, userID)

	// Register client
	h.hub.Register(userID, client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
