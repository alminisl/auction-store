import { PublicUser } from './user';

export interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  created_at: string;
}

export interface MessageWithSender extends Message {
  sender: PublicUser;
}

export interface Conversation {
  id: string;
  other_user: PublicUser;
  last_message?: Message;
  last_message_at?: string;
  unread_count: number;
  created_at: string;
}

export interface SendMessageRequest {
  recipient_id: string;
  content: string;
}

export interface SendMessageResponse {
  message: Message;
  conversation_id: string;
}

export interface ConversationsResponse {
  conversations: Conversation[];
}

export interface MessagesResponse {
  messages: MessageWithSender[];
}

export interface UnreadCountResponse {
  count: number;
}

export type WSMessageType = 'new_message' | 'message_read' | 'typing_started' | 'typing_stopped';

export interface WSMessagePayload {
  type: WSMessageType;
  message?: Message;
  conversation_id?: string;
  sender_id?: string;
}
