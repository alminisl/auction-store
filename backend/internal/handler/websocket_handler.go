package handler

import (
	"log"
	"net/http"

	"github.com/auction-cards/backend/internal/middleware"
	ws "github.com/auction-cards/backend/internal/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// In production, check against allowed origins
		return true
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) HandleAuctionWS(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "id")
	if err != nil {
		http.Error(w, "Invalid auction ID", http.StatusBadRequest)
		return
	}

	// Get user ID if authenticated (optional)
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		userID = uuid.New() // Generate anonymous ID for non-authenticated users
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := ws.NewClient(h.hub, conn, auctionID, userID)

	// Register client
	h.hub.Register(auctionID, client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}
