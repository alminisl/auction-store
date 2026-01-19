package domain

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat between two users
type Conversation struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ParticipantOne uuid.UUID  `json:"participant_one" db:"participant_one"`
	ParticipantTwo uuid.UUID  `json:"participant_two" db:"participant_two"`
	LastMessageAt  *time.Time `json:"last_message_at" db:"last_message_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ConversationWithDetails includes participant info and unread count
type ConversationWithDetails struct {
	ID              uuid.UUID   `json:"id"`
	OtherUser       *PublicUser `json:"other_user"`
	LastMessage     *Message    `json:"last_message,omitempty"`
	LastMessageAt   *time.Time  `json:"last_message_at"`
	UnreadCount     int         `json:"unread_count"`
	CreatedAt       time.Time   `json:"created_at"`
}

// Message represents a single message in a conversation
type Message struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ConversationID   uuid.UUID `json:"conversation_id" db:"conversation_id"`
	SenderID         uuid.UUID `json:"sender_id" db:"sender_id"`
	ContentEncrypted []byte    `json:"-" db:"content_encrypted"`
	ContentNonce     []byte    `json:"-" db:"content_nonce"`
	Content          string    `json:"content" db:"-"` // Decrypted content, not stored in DB
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// MessageWithSender includes sender info
type MessageWithSender struct {
	Message
	Sender *PublicUser `json:"sender"`
}

// ConversationReadStatus tracks when a user last read a conversation
type ConversationReadStatus struct {
	ConversationID uuid.UUID `json:"conversation_id" db:"conversation_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	LastReadAt     time.Time `json:"last_read_at" db:"last_read_at"`
}

// Request DTOs
type SendMessageRequest struct {
	RecipientID uuid.UUID `json:"recipient_id" validate:"required"`
	Content     string    `json:"content" validate:"required,min=1,max=5000"`
}

type GetMessagesRequest struct {
	Page  int `json:"page" validate:"omitempty,min=1"`
	Limit int `json:"limit" validate:"omitempty,min=1,max=100"`
}

// Response DTOs
type SendMessageResponse struct {
	Message        *Message  `json:"message"`
	ConversationID uuid.UUID `json:"conversation_id"`
}

type ConversationsResponse struct {
	Conversations []ConversationWithDetails `json:"conversations"`
}

type MessagesResponse struct {
	Messages []MessageWithSender `json:"messages"`
}

type UnreadCountResponse struct {
	Count int `json:"count"`
}

// WebSocket message types for real-time messaging
type MessageWSType string

const (
	MessageWSTypeNewMessage    MessageWSType = "new_message"
	MessageWSTypeMessageRead   MessageWSType = "message_read"
	MessageWSTypeTypingStarted MessageWSType = "typing_started"
	MessageWSTypeTypingStopped MessageWSType = "typing_stopped"
)

type MessageWSPayload struct {
	Type           MessageWSType `json:"type"`
	Message        *Message      `json:"message,omitempty"`
	ConversationID uuid.UUID     `json:"conversation_id,omitempty"`
	SenderID       uuid.UUID     `json:"sender_id,omitempty"`
}
