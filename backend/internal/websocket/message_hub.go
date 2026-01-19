package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/auction-cards/backend/internal/cache"
	"github.com/google/uuid"
)

// MessageHub manages WebSocket connections for messaging
type MessageHub struct {
	// Registered clients by user ID (one user can have multiple connections)
	users map[uuid.UUID]map[*MessageClient]bool

	// Register requests
	register chan *messageSubscription

	// Unregister requests
	unregister chan *messageSubscription

	// Send message to specific user
	sendToUser chan *userMessage

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Redis cache for pub/sub
	redis *cache.RedisCache

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

type messageSubscription struct {
	userID uuid.UUID
	client *MessageClient
}

type userMessage struct {
	userID  uuid.UUID
	message []byte
}

func NewMessageHub(redis *cache.RedisCache) *MessageHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &MessageHub{
		users:      make(map[uuid.UUID]map[*MessageClient]bool),
		register:   make(chan *messageSubscription),
		unregister: make(chan *messageSubscription),
		sendToUser: make(chan *userMessage, 256),
		redis:      redis,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (h *MessageHub) Run() {
	// Start Redis subscriber
	if h.redis != nil {
		go h.subscribeToRedis()
	}

	for {
		select {
		case <-h.ctx.Done():
			return

		case sub := <-h.register:
			h.mu.Lock()
			if h.users[sub.userID] == nil {
				h.users[sub.userID] = make(map[*MessageClient]bool)
			}
			h.users[sub.userID][sub.client] = true
			h.mu.Unlock()
			log.Printf("Message client registered for user %s", sub.userID)

		case sub := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.users[sub.userID]; ok {
				if _, ok := clients[sub.client]; ok {
					delete(clients, sub.client)
					close(sub.client.send)
					if len(clients) == 0 {
						delete(h.users, sub.userID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Message client unregistered for user %s", sub.userID)

		case msg := <-h.sendToUser:
			h.mu.RLock()
			if clients, ok := h.users[msg.userID]; ok {
				for client := range clients {
					select {
					case client.send <- msg.message:
					default:
						// Client's buffer is full, close connection
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *MessageHub) Stop() {
	h.cancel()
}

func (h *MessageHub) Register(userID uuid.UUID, client *MessageClient) {
	h.register <- &messageSubscription{userID: userID, client: client}
}

func (h *MessageHub) Unregister(userID uuid.UUID, client *MessageClient) {
	h.unregister <- &messageSubscription{userID: userID, client: client}
}

// SendToUser sends a message to all connections of a specific user
func (h *MessageHub) SendToUser(userID uuid.UUID, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.sendToUser <- &userMessage{
		userID:  userID,
		message: data,
	}

	// Also publish to Redis for cross-instance delivery
	if h.redis != nil {
		channel := "message:" + userID.String()
		h.redis.Client().Publish(h.ctx, channel, string(data))
	}
}

func (h *MessageHub) subscribeToRedis() {
	// Subscribe to all message channels using pattern
	pubsub := h.redis.Client().PSubscribe(h.ctx, "message:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-h.ctx.Done():
			return
		case msg := <-ch:
			// Extract user ID from channel name (message:{uuid})
			channelName := msg.Channel
			if len(channelName) > 8 {
				userIDStr := channelName[8:] // Remove "message:" prefix
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					continue
				}

				// Only deliver if we have local connections for this user
				h.mu.RLock()
				if clients, ok := h.users[userID]; ok {
					for client := range clients {
						select {
						case client.send <- []byte(msg.Payload):
						default:
							// Client's buffer is full
						}
					}
				}
				h.mu.RUnlock()
			}
		}
	}
}

// IsUserOnline checks if a user has any active WebSocket connections
func (h *MessageHub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.users[userID]
	return ok && len(clients) > 0
}

// GetOnlineUserCount returns the number of users with active connections
func (h *MessageHub) GetOnlineUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.users)
}
