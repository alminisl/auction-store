package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/auction-cards/backend/internal/cache"
	"github.com/google/uuid"
)

type Hub struct {
	// Registered clients by auction ID
	auctions map[uuid.UUID]map[*Client]bool

	// Register requests
	register chan *subscription

	// Unregister requests
	unregister chan *subscription

	// Broadcast to auction
	broadcast chan *auctionMessage

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Redis cache for pub/sub
	redis *cache.RedisCache

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

type subscription struct {
	auctionID uuid.UUID
	client    *Client
}

type auctionMessage struct {
	auctionID uuid.UUID
	message   []byte
}

func NewHub(redis *cache.RedisCache) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		auctions:   make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *subscription),
		unregister: make(chan *subscription),
		broadcast:  make(chan *auctionMessage, 256),
		redis:      redis,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (h *Hub) Run() {
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
			if h.auctions[sub.auctionID] == nil {
				h.auctions[sub.auctionID] = make(map[*Client]bool)
			}
			h.auctions[sub.auctionID][sub.client] = true
			h.mu.Unlock()
			log.Printf("Client registered for auction %s", sub.auctionID)

		case sub := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.auctions[sub.auctionID]; ok {
				if _, ok := clients[sub.client]; ok {
					delete(clients, sub.client)
					close(sub.client.send)
					if len(clients) == 0 {
						delete(h.auctions, sub.auctionID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client unregistered from auction %s", sub.auctionID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.auctions[msg.auctionID]; ok {
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

func (h *Hub) Stop() {
	h.cancel()
}

func (h *Hub) Register(auctionID uuid.UUID, client *Client) {
	h.register <- &subscription{auctionID: auctionID, client: client}
}

func (h *Hub) Unregister(auctionID uuid.UUID, client *Client) {
	h.unregister <- &subscription{auctionID: auctionID, client: client}
}

func (h *Hub) BroadcastToAuction(auctionID uuid.UUID, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.broadcast <- &auctionMessage{
		auctionID: auctionID,
		message:   data,
	}
}

func (h *Hub) subscribeToRedis() {
	// Subscribe to all auction channels using pattern
	pubsub := h.redis.Client().PSubscribe(h.ctx, "auction:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-h.ctx.Done():
			return
		case msg := <-ch:
			// Extract auction ID from channel name (auction:{uuid})
			channelName := msg.Channel
			if len(channelName) > 8 {
				auctionIDStr := channelName[8:] // Remove "auction:" prefix
				auctionID, err := uuid.Parse(auctionIDStr)
				if err != nil {
					continue
				}

				h.broadcast <- &auctionMessage{
					auctionID: auctionID,
					message:   []byte(msg.Payload),
				}
			}
		}
	}
}

func (h *Hub) GetClientCount(auctionID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.auctions[auctionID]; ok {
		return len(clients)
	}
	return 0
}
