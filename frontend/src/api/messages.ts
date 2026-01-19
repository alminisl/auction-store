import api from './client';
import {
  APIResponse,
  SendMessageRequest,
  SendMessageResponse,
  ConversationsResponse,
  MessagesResponse,
  UnreadCountResponse,
  Conversation,
} from '../types';

export const messagesApi = {
  async sendMessage(data: SendMessageRequest): Promise<APIResponse<SendMessageResponse>> {
    const response = await api.post<APIResponse<SendMessageResponse>>('/messages', data);
    return response.data;
  },

  async getUnreadCount(): Promise<APIResponse<UnreadCountResponse>> {
    const response = await api.get<APIResponse<UnreadCountResponse>>('/messages/unread-count');
    return response.data;
  },

  async getConversations(): Promise<APIResponse<ConversationsResponse>> {
    const response = await api.get<APIResponse<ConversationsResponse>>('/conversations');
    return response.data;
  },

  async getConversation(id: string): Promise<APIResponse<Conversation>> {
    const response = await api.get<APIResponse<Conversation>>(`/conversations/${id}`);
    return response.data;
  },

  async getMessages(
    conversationId: string,
    params?: { page?: number; limit?: number }
  ): Promise<APIResponse<MessagesResponse>> {
    const response = await api.get<APIResponse<MessagesResponse>>(
      `/conversations/${conversationId}/messages`,
      { params }
    );
    return response.data;
  },

  async markAsRead(conversationId: string): Promise<APIResponse<void>> {
    const response = await api.put<APIResponse<void>>(`/conversations/${conversationId}/read`);
    return response.data;
  },
};

export default messagesApi;
