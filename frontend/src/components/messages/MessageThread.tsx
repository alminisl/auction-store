import { useEffect, useRef } from 'react';
import { Loader2, ArrowLeft, User } from 'lucide-react';
import { Conversation, MessageWithSender } from '../../types';
import { MessageBubble } from './MessageBubble';
import { MessageInput } from './MessageInput';

interface MessageThreadProps {
  conversation: Conversation | null;
  messages: MessageWithSender[];
  currentUserId: string;
  isLoading: boolean;
  isSending: boolean;
  onSendMessage: (content: string) => Promise<void>;
  onBack?: () => void;
  showBackButton?: boolean;
}

export function MessageThread({
  conversation,
  messages,
  currentUserId,
  isLoading,
  isSending,
  onSendMessage,
  onBack,
  showBackButton = false,
}: MessageThreadProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  if (!conversation) {
    return (
      <div className="flex-1 flex items-center justify-center bg-muted/30">
        <div className="text-center text-muted-foreground">
          <p>Select a conversation to start messaging</p>
        </div>
      </div>
    );
  }

  const otherUser = conversation.other_user;

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="border-b border-border p-4 flex items-center gap-3">
        {showBackButton && onBack && (
          <button
            onClick={onBack}
            className="p-1 rounded hover:bg-muted transition-colors md:hidden"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
        )}
        <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center overflow-hidden">
          {otherUser.avatar_url ? (
            <img
              src={otherUser.avatar_url}
              alt={otherUser.username}
              className="w-full h-full object-cover"
            />
          ) : (
            <User className="h-5 w-5 text-muted-foreground" />
          )}
        </div>
        <div>
          <h3 className="font-medium">{otherUser.username}</h3>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {isLoading && (
          <div className="flex justify-center py-4">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        )}
        {!isLoading && messages.length === 0 && (
          <div className="text-center text-muted-foreground py-8">
            <p>No messages yet. Start the conversation!</p>
          </div>
        )}
        {messages.map((message) => (
          <MessageBubble
            key={message.id}
            message={message}
            isOwn={message.sender_id === currentUserId}
          />
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <MessageInput
        onSend={onSendMessage}
        isSending={isSending}
      />
    </div>
  );
}

export default MessageThread;
