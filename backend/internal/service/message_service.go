package service

import (
	"context"
	"fmt"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/pkg/encryption"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/auction-cards/backend/internal/websocket"
	"github.com/google/uuid"
)

type MessageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	encryptor   *encryption.AESEncryptor
	messageHub  *websocket.MessageHub
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	encryptionKey string,
	messageHub *websocket.MessageHub,
) (*MessageService, error) {
	encryptor, err := encryption.NewAESEncryptor(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryptor: %w", err)
	}

	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		encryptor:   encryptor,
		messageHub:  messageHub,
	}, nil
}

// SendMessage sends a message from one user to another
func (s *MessageService) SendMessage(ctx context.Context, senderID uuid.UUID, req *domain.SendMessageRequest) (*domain.Message, uuid.UUID, error) {
	// Check that recipient exists
	recipient, err := s.userRepo.GetByID(ctx, req.RecipientID)
	if err != nil {
		return nil, uuid.Nil, domain.ErrNotFound
	}

	// Cannot message yourself
	if senderID == recipient.ID {
		return nil, uuid.Nil, domain.ErrValidation
	}

	// Get or create conversation
	conv, err := s.messageRepo.GetOrCreateConversation(ctx, senderID, req.RecipientID)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to get/create conversation: %w", err)
	}

	// Encrypt the message content
	ciphertext, nonce, err := s.encryptor.EncryptString(req.Content)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	// Create the message
	msg := &domain.Message{
		ConversationID:   conv.ID,
		SenderID:         senderID,
		ContentEncrypted: ciphertext,
		ContentNonce:     nonce,
		Content:          req.Content, // Keep plaintext in memory for response
	}

	if err := s.messageRepo.CreateMessage(ctx, msg); err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Send real-time notification to recipient via WebSocket
	if s.messageHub != nil {
		wsPayload := domain.MessageWSPayload{
			Type:           domain.MessageWSTypeNewMessage,
			Message:        msg,
			ConversationID: conv.ID,
			SenderID:       senderID,
		}
		s.messageHub.SendToUser(req.RecipientID, wsPayload)
	}

	return msg, conv.ID, nil
}

// GetConversations returns all conversations for a user with details
func (s *MessageService) GetConversations(ctx context.Context, userID uuid.UUID) ([]domain.ConversationWithDetails, error) {
	conversations, err := s.messageRepo.GetConversationsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	result := make([]domain.ConversationWithDetails, 0, len(conversations))

	for _, conv := range conversations {
		// Determine the other user
		otherUserID := conv.ParticipantOne
		if otherUserID == userID {
			otherUserID = conv.ParticipantTwo
		}

		// Get other user's info
		otherUser, err := s.userRepo.GetByID(ctx, otherUserID)
		if err != nil {
			continue // Skip if user not found
		}

		// Get unread count
		unreadCount, _ := s.messageRepo.GetUnreadCountForConversation(ctx, conv.ID, userID)

		// Get last message
		var lastMsg *domain.Message
		lastMsgRaw, _ := s.messageRepo.GetLastMessage(ctx, conv.ID)
		if lastMsgRaw != nil {
			// Decrypt the message
			plaintext, err := s.encryptor.DecryptString(lastMsgRaw.ContentEncrypted, lastMsgRaw.ContentNonce)
			if err == nil {
				lastMsgRaw.Content = plaintext
				lastMsg = lastMsgRaw
			}
		}

		result = append(result, domain.ConversationWithDetails{
			ID:            conv.ID,
			OtherUser:     otherUser.ToPublic(),
			LastMessage:   lastMsg,
			LastMessageAt: conv.LastMessageAt,
			UnreadCount:   unreadCount,
			CreatedAt:     conv.CreatedAt,
		})
	}

	return result, nil
}

// GetMessages returns messages for a conversation
func (s *MessageService) GetMessages(ctx context.Context, userID, conversationID uuid.UUID, page, limit int) ([]domain.MessageWithSender, int, error) {
	// Verify user is a participant
	isMember, err := s.messageRepo.IsUserInConversation(ctx, conversationID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, 0, domain.ErrForbidden
	}

	messages, totalCount, err := s.messageRepo.GetMessagesByConversation(ctx, conversationID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get messages: %w", err)
	}

	// Get conversation to find participants
	conv, err := s.messageRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Cache users
	users := make(map[uuid.UUID]*domain.PublicUser)
	for _, uid := range []uuid.UUID{conv.ParticipantOne, conv.ParticipantTwo} {
		user, err := s.userRepo.GetByID(ctx, uid)
		if err == nil {
			users[uid] = user.ToPublic()
		}
	}

	result := make([]domain.MessageWithSender, 0, len(messages))
	for _, msg := range messages {
		// Decrypt message content
		plaintext, err := s.encryptor.DecryptString(msg.ContentEncrypted, msg.ContentNonce)
		if err != nil {
			continue // Skip messages that can't be decrypted
		}
		msg.Content = plaintext

		result = append(result, domain.MessageWithSender{
			Message: msg,
			Sender:  users[msg.SenderID],
		})
	}

	return result, totalCount, nil
}

// MarkConversationRead marks all messages in a conversation as read
func (s *MessageService) MarkConversationRead(ctx context.Context, userID, conversationID uuid.UUID) error {
	// Verify user is a participant
	isMember, err := s.messageRepo.IsUserInConversation(ctx, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return domain.ErrForbidden
	}

	return s.messageRepo.UpdateReadStatus(ctx, conversationID, userID)
}

// GetUnreadCount returns the total unread message count for a user
func (s *MessageService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.messageRepo.GetTotalUnreadCount(ctx, userID)
}

// GetConversationByID retrieves a single conversation with details
func (s *MessageService) GetConversationByID(ctx context.Context, userID, conversationID uuid.UUID) (*domain.ConversationWithDetails, error) {
	// Verify user is a participant
	isMember, err := s.messageRepo.IsUserInConversation(ctx, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, domain.ErrForbidden
	}

	conv, err := s.messageRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Determine the other user
	otherUserID := conv.ParticipantOne
	if otherUserID == userID {
		otherUserID = conv.ParticipantTwo
	}

	// Get other user's info
	otherUser, err := s.userRepo.GetByID(ctx, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get other user: %w", err)
	}

	// Get unread count
	unreadCount, _ := s.messageRepo.GetUnreadCountForConversation(ctx, conv.ID, userID)

	// Get last message
	var lastMsg *domain.Message
	lastMsgRaw, _ := s.messageRepo.GetLastMessage(ctx, conv.ID)
	if lastMsgRaw != nil {
		plaintext, err := s.encryptor.DecryptString(lastMsgRaw.ContentEncrypted, lastMsgRaw.ContentNonce)
		if err == nil {
			lastMsgRaw.Content = plaintext
			lastMsg = lastMsgRaw
		}
	}

	return &domain.ConversationWithDetails{
		ID:            conv.ID,
		OtherUser:     otherUser.ToPublic(),
		LastMessage:   lastMsg,
		LastMessageAt: conv.LastMessageAt,
		UnreadCount:   unreadCount,
		CreatedAt:     conv.CreatedAt,
	}, nil
}
