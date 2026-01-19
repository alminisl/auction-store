package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MessageRepository struct {
	db *DB
}

func NewMessageRepository(db *DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// GetOrCreateConversation returns an existing conversation or creates a new one
func (r *MessageRepository) GetOrCreateConversation(ctx context.Context, userOne, userTwo uuid.UUID) (*domain.Conversation, error) {
	// Ensure consistent ordering (participant_one < participant_two)
	participantOne, participantTwo := userOne, userTwo
	if participantOne.String() > participantTwo.String() {
		participantOne, participantTwo = participantTwo, participantOne
	}

	q := r.db.GetQuerier(ctx)

	// Try to get existing conversation
	conv := &domain.Conversation{}
	query := `
		SELECT id, participant_one, participant_two, last_message_at, created_at
		FROM conversations
		WHERE participant_one = $1 AND participant_two = $2`

	err := q.QueryRow(ctx, query, participantOne, participantTwo).Scan(
		&conv.ID,
		&conv.ParticipantOne,
		&conv.ParticipantTwo,
		&conv.LastMessageAt,
		&conv.CreatedAt,
	)

	if err == nil {
		return conv, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Create new conversation
	conv = &domain.Conversation{
		ID:             uuid.New(),
		ParticipantOne: participantOne,
		ParticipantTwo: participantTwo,
	}

	insertQuery := `
		INSERT INTO conversations (id, participant_one, participant_two)
		VALUES ($1, $2, $3)
		RETURNING created_at`

	err = q.QueryRow(ctx, insertQuery, conv.ID, conv.ParticipantOne, conv.ParticipantTwo).Scan(&conv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv, nil
}

// GetConversationByID retrieves a conversation by ID
func (r *MessageRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	query := `
		SELECT id, participant_one, participant_two, last_message_at, created_at
		FROM conversations
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	conv := &domain.Conversation{}
	err := q.QueryRow(ctx, query, id).Scan(
		&conv.ID,
		&conv.ParticipantOne,
		&conv.ParticipantTwo,
		&conv.LastMessageAt,
		&conv.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conv, nil
}

// GetConversationsForUser retrieves all conversations for a user
func (r *MessageRepository) GetConversationsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
	query := `
		SELECT id, participant_one, participant_two, last_message_at, created_at
		FROM conversations
		WHERE participant_one = $1 OR participant_two = $1
		ORDER BY COALESCE(last_message_at, created_at) DESC`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]domain.Conversation, 0)
	for rows.Next() {
		var conv domain.Conversation
		err := rows.Scan(
			&conv.ID,
			&conv.ParticipantOne,
			&conv.ParticipantTwo,
			&conv.LastMessageAt,
			&conv.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// CreateMessage creates a new message
func (r *MessageRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}

	query := `
		INSERT INTO messages (id, conversation_id, sender_id, content_encrypted, content_nonce)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		msg.ID,
		msg.ConversationID,
		msg.SenderID,
		msg.ContentEncrypted,
		msg.ContentNonce,
	).Scan(&msg.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation's last_message_at
	updateQuery := `UPDATE conversations SET last_message_at = $2 WHERE id = $1`
	_, err = q.Exec(ctx, updateQuery, msg.ConversationID, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// GetMessagesByConversation retrieves messages for a conversation
func (r *MessageRepository) GetMessagesByConversation(ctx context.Context, conversationID uuid.UUID, page, limit int) ([]domain.Message, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}

	countQuery := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`

	q := r.db.GetQuerier(ctx)
	var totalCount int
	if err := q.QueryRow(ctx, countQuery, conversationID).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, conversation_id, sender_id, content_encrypted, content_nonce, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := q.Query(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.ContentEncrypted,
			&msg.ContentNonce,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, totalCount, nil
}

// GetLastMessage retrieves the last message in a conversation
func (r *MessageRepository) GetLastMessage(ctx context.Context, conversationID uuid.UUID) (*domain.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content_encrypted, content_nonce, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	q := r.db.GetQuerier(ctx)
	msg := &domain.Message{}
	err := q.QueryRow(ctx, query, conversationID).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.ContentEncrypted,
		&msg.ContentNonce,
		&msg.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // No messages yet
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}

	return msg, nil
}

// UpdateReadStatus updates the read status for a user in a conversation
func (r *MessageRepository) UpdateReadStatus(ctx context.Context, conversationID, userID uuid.UUID) error {
	query := `
		INSERT INTO conversation_read_status (conversation_id, user_id, last_read_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (conversation_id, user_id)
		DO UPDATE SET last_read_at = EXCLUDED.last_read_at`

	q := r.db.GetQuerier(ctx)
	_, err := q.Exec(ctx, query, conversationID, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update read status: %w", err)
	}

	return nil
}

// GetReadStatus retrieves the read status for a user in a conversation
func (r *MessageRepository) GetReadStatus(ctx context.Context, conversationID, userID uuid.UUID) (*domain.ConversationReadStatus, error) {
	query := `
		SELECT conversation_id, user_id, last_read_at
		FROM conversation_read_status
		WHERE conversation_id = $1 AND user_id = $2`

	q := r.db.GetQuerier(ctx)
	status := &domain.ConversationReadStatus{}
	err := q.QueryRow(ctx, query, conversationID, userID).Scan(
		&status.ConversationID,
		&status.UserID,
		&status.LastReadAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Never read
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get read status: %w", err)
	}

	return status, nil
}

// GetUnreadCountForConversation returns the count of unread messages in a conversation
func (r *MessageRepository) GetUnreadCountForConversation(ctx context.Context, conversationID, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		WHERE m.conversation_id = $1
			AND m.sender_id != $2
			AND m.created_at > COALESCE(
				(SELECT last_read_at FROM conversation_read_status WHERE conversation_id = $1 AND user_id = $2),
				'1970-01-01'::timestamp
			)`

	q := r.db.GetQuerier(ctx)
	var count int
	if err := q.QueryRow(ctx, query, conversationID, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// GetTotalUnreadCount returns the total count of unread messages across all conversations
func (r *MessageRepository) GetTotalUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE (c.participant_one = $1 OR c.participant_two = $1)
			AND m.sender_id != $1
			AND m.created_at > COALESCE(
				(SELECT last_read_at FROM conversation_read_status WHERE conversation_id = m.conversation_id AND user_id = $1),
				'1970-01-01'::timestamp
			)`

	q := r.db.GetQuerier(ctx)
	var count int
	if err := q.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get total unread count: %w", err)
	}

	return count, nil
}

// IsUserInConversation checks if a user is a participant in a conversation
func (r *MessageRepository) IsUserInConversation(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM conversations
			WHERE id = $1 AND (participant_one = $2 OR participant_two = $2)
		)`

	q := r.db.GetQuerier(ctx)
	var exists bool
	if err := q.QueryRow(ctx, query, conversationID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check conversation membership: %w", err)
	}

	return exists, nil
}
