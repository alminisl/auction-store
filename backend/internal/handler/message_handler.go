package handler

import (
	"net/http"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// SendMessage handles POST /api/messages
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req domain.SendMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	userID := getUserID(r)
	msg, conversationID, err := h.messageService.SendMessage(r.Context(), userID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, &domain.SendMessageResponse{
		Message:        msg,
		ConversationID: conversationID,
	})
}

// GetConversations handles GET /api/conversations
func (h *MessageHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	conversations, err := h.messageService.GetConversations(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, &domain.ConversationsResponse{
		Conversations: conversations,
	})
}

// GetMessages handles GET /api/conversations/{id}/messages
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	conversationID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid conversation ID")
		return
	}

	userID := getUserID(r)
	page := getQueryParamInt(r, "page", 1)
	limit := getQueryParamInt(r, "limit", 50)

	messages, totalCount, err := h.messageService.GetMessages(r.Context(), userID, conversationID, page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	totalPages := (totalCount + limit - 1) / limit

	respondJSONWithMeta(w, http.StatusOK, &domain.MessagesResponse{
		Messages: messages,
	}, &domain.APIMeta{
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

// MarkAsRead handles PUT /api/conversations/{id}/read
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	conversationID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid conversation ID")
		return
	}

	userID := getUserID(r)

	if err := h.messageService.MarkConversationRead(r.Context(), userID, conversationID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Conversation marked as read",
	})
}

// GetUnreadCount handles GET /api/messages/unread-count
func (h *MessageHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	count, err := h.messageService.GetUnreadCount(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, &domain.UnreadCountResponse{
		Count: count,
	})
}

// GetConversation handles GET /api/conversations/{id}
func (h *MessageHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	conversationID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid conversation ID")
		return
	}

	userID := getUserID(r)

	conversation, err := h.messageService.GetConversationByID(r.Context(), userID, conversationID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, conversation)
}
