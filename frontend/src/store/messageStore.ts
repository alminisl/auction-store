import { create } from 'zustand';
import { Conversation, Message, MessageWithSender } from '../types';
import { messagesApi } from '../api/messages';

interface MessageState {
  conversations: Conversation[];
  currentConversation: Conversation | null;
  messages: MessageWithSender[];
  unreadCount: number;
  isLoading: boolean;
  isSending: boolean;

  // Actions
  fetchConversations: () => Promise<void>;
  fetchMessages: (conversationId: string, page?: number) => Promise<void>;
  fetchUnreadCount: () => Promise<void>;
  sendMessage: (recipientId: string, content: string) => Promise<Message | null>;
  markAsRead: (conversationId: string) => Promise<void>;
  setCurrentConversation: (conversation: Conversation | null) => void;
  addIncomingMessage: (message: Message, conversationId: string) => void;
  clearMessages: () => void;
}

export const useMessageStore = create<MessageState>((set, get) => ({
  conversations: [],
  currentConversation: null,
  messages: [],
  unreadCount: 0,
  isLoading: false,
  isSending: false,

  fetchConversations: async () => {
    set({ isLoading: true });
    try {
      const response = await messagesApi.getConversations();
      if (response.success && response.data) {
        set({ conversations: response.data.conversations, isLoading: false });
        // Also update unread count
        const totalUnread = response.data.conversations.reduce(
          (sum, conv) => sum + conv.unread_count,
          0
        );
        set({ unreadCount: totalUnread });
      }
    } catch {
      set({ isLoading: false });
    }
  },

  fetchMessages: async (conversationId: string, page = 1) => {
    set({ isLoading: true });
    try {
      const response = await messagesApi.getMessages(conversationId, { page, limit: 50 });
      if (response.success && response.data) {
        // Messages come in DESC order (newest first), reverse for display
        const newMessages = [...response.data.messages].reverse();
        if (page === 1) {
          set({ messages: newMessages, isLoading: false });
        } else {
          // Prepend older messages
          set({ messages: [...newMessages, ...get().messages], isLoading: false });
        }
      }
    } catch {
      set({ isLoading: false });
    }
  },

  fetchUnreadCount: async () => {
    try {
      const response = await messagesApi.getUnreadCount();
      if (response.success && response.data) {
        set({ unreadCount: response.data.count });
      }
    } catch {
      // Ignore errors
    }
  },

  sendMessage: async (recipientId: string, content: string) => {
    set({ isSending: true });
    try {
      const response = await messagesApi.sendMessage({ recipient_id: recipientId, content });
      if (response.success && response.data) {
        const { message, conversation_id } = response.data;

        // Add message to current messages list
        const currentMessages = get().messages;
        const messageWithSender: MessageWithSender = {
          ...message,
          sender: get().currentConversation?.other_user || ({} as MessageWithSender['sender']),
        };
        set({ messages: [...currentMessages, messageWithSender], isSending: false });

        // Update conversation in list
        const conversations = get().conversations.map((conv) => {
          if (conv.id === conversation_id) {
            return {
              ...conv,
              last_message: message,
              last_message_at: message.created_at,
            };
          }
          return conv;
        });

        // Sort conversations by last message time
        conversations.sort((a, b) => {
          const aTime = a.last_message_at || a.created_at;
          const bTime = b.last_message_at || b.created_at;
          return new Date(bTime).getTime() - new Date(aTime).getTime();
        });

        set({ conversations });
        return message;
      }
      set({ isSending: false });
      return null;
    } catch {
      set({ isSending: false });
      return null;
    }
  },

  markAsRead: async (conversationId: string) => {
    try {
      await messagesApi.markAsRead(conversationId);
      // Update conversation unread count
      const conversations = get().conversations.map((conv) => {
        if (conv.id === conversationId) {
          return { ...conv, unread_count: 0 };
        }
        return conv;
      });
      set({ conversations });
      // Recalculate total unread
      const totalUnread = conversations.reduce((sum, conv) => sum + conv.unread_count, 0);
      set({ unreadCount: totalUnread });
    } catch {
      // Ignore errors
    }
  },

  setCurrentConversation: (conversation: Conversation | null) => {
    set({ currentConversation: conversation, messages: [] });
  },

  addIncomingMessage: (message: Message, conversationId: string) => {
    const currentConv = get().currentConversation;

    // If this is for the current conversation, add to messages
    if (currentConv && currentConv.id === conversationId) {
      // Find sender from current conversation
      const sender = currentConv.other_user;
      const messageWithSender: MessageWithSender = {
        ...message,
        sender,
      };
      set({ messages: [...get().messages, messageWithSender] });
    }

    // Update conversation list
    const conversations = get().conversations.map((conv) => {
      if (conv.id === conversationId) {
        return {
          ...conv,
          last_message: message,
          last_message_at: message.created_at,
          unread_count: currentConv?.id === conversationId ? conv.unread_count : conv.unread_count + 1,
        };
      }
      return conv;
    });

    // If conversation not found, we need to refetch
    if (!conversations.find((c) => c.id === conversationId)) {
      get().fetchConversations();
    } else {
      // Sort conversations by last message time
      conversations.sort((a, b) => {
        const aTime = a.last_message_at || a.created_at;
        const bTime = b.last_message_at || b.created_at;
        return new Date(bTime).getTime() - new Date(aTime).getTime();
      });
      set({ conversations });

      // Update unread count if not in current conversation
      if (!currentConv || currentConv.id !== conversationId) {
        set({ unreadCount: get().unreadCount + 1 });
      }
    }
  },

  clearMessages: () => {
    set({ messages: [], currentConversation: null });
  },
}));

export default useMessageStore;
