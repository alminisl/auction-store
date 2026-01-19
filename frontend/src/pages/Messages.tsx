import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { MessageSquare, Loader2 } from 'lucide-react';
import { useAuthStore, useMessageStore } from '../store';
import { useMessageWebSocket } from '../hooks';
import { ConversationList } from '../components/messages/ConversationList';
import { MessageThread } from '../components/messages/MessageThread';
import { Conversation, PublicUser } from '../types';
import { cn } from '../utils/cn';
import { usersApi } from '../api';

export default function Messages() {
  const { user } = useAuthStore();
  const [searchParams, setSearchParams] = useSearchParams();
  const [isMobileViewingThread, setIsMobileViewingThread] = useState(false);
  const [pendingRecipient, setPendingRecipient] = useState<PublicUser | null>(null);

  const {
    conversations,
    currentConversation,
    messages,
    isLoading,
    isSending,
    fetchConversations,
    fetchMessages,
    sendMessage,
    markAsRead,
    setCurrentConversation,
  } = useMessageStore();

  // Connect to WebSocket for real-time updates
  useMessageWebSocket({ enabled: !!user });

  // Fetch conversations on mount
  useEffect(() => {
    fetchConversations();
  }, [fetchConversations]);

  // Handle recipient from URL params (e.g., from auction page "Message Seller")
  useEffect(() => {
    const recipientId = searchParams.get('recipient');
    if (recipientId && conversations.length >= 0 && !isLoading) {
      // Check if we already have a conversation with this user
      const existingConv = conversations.find(
        (c) => c.other_user.id === recipientId
      );

      if (existingConv) {
        handleSelectConversation(existingConv);
        // Clear the recipient param
        setSearchParams({});
      } else {
        // Fetch the user info to start a new conversation
        usersApi.getUser(recipientId).then((response) => {
          if (response.success && response.data) {
            const userData = response.data as unknown as { user: PublicUser };
            setPendingRecipient(userData.user || response.data);
            setIsMobileViewingThread(true);
            // Clear the recipient param
            setSearchParams({});
          }
        }).catch(() => {
          // User not found, ignore
        });
      }
    }
  }, [searchParams, conversations, isLoading]);

  // Handle conversation selection from URL params (e.g., from user profile)
  useEffect(() => {
    const conversationId = searchParams.get('conversation');
    if (conversationId && conversations.length > 0) {
      const conversation = conversations.find((c) => c.id === conversationId);
      if (conversation) {
        handleSelectConversation(conversation);
      }
    }
  }, [searchParams, conversations]);

  const handleSelectConversation = async (conversation: Conversation) => {
    setPendingRecipient(null);
    setCurrentConversation(conversation);
    setIsMobileViewingThread(true);
    await fetchMessages(conversation.id);
    if (conversation.unread_count > 0) {
      await markAsRead(conversation.id);
    }
  };

  const handleSendMessage = async (content: string) => {
    // If we have a pending recipient (new conversation), send to them
    if (pendingRecipient) {
      const message = await sendMessage(pendingRecipient.id, content);
      if (message) {
        // Refresh conversations to get the new one
        await fetchConversations();
        setPendingRecipient(null);
      }
      return;
    }

    if (!currentConversation) return;
    await sendMessage(currentConversation.other_user.id, content);
  };

  const handleBackToList = () => {
    setIsMobileViewingThread(false);
    setCurrentConversation(null);
    setPendingRecipient(null);
  };

  return (
    <div className="container-custom py-4 md:py-8">
      {/* Header */}
      <div className="mb-4 md:mb-6">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <MessageSquare className="h-6 w-6 text-primary" />
          Messages
        </h1>
      </div>

      {/* Main Content */}
      <div className="bg-card border border-border rounded-lg overflow-hidden h-[calc(100vh-180px)] md:h-[calc(100vh-220px)]">
        <div className="flex h-full">
          {/* Conversation List (hide on mobile when viewing thread) */}
          <div
            className={cn(
              'w-full md:w-80 lg:w-96 border-r border-border flex flex-col',
              isMobileViewingThread ? 'hidden md:flex' : 'flex'
            )}
          >
            <div className="p-4 border-b border-border">
              <h2 className="font-medium">Conversations</h2>
            </div>
            {isLoading && conversations.length === 0 ? (
              <div className="flex-1 flex items-center justify-center">
                <Loader2 className="h-6 w-6 animate-spin text-primary" />
              </div>
            ) : (
              <ConversationList
                conversations={conversations}
                selectedId={currentConversation?.id}
                onSelect={handleSelectConversation}
              />
            )}
          </div>

          {/* Message Thread (show on mobile only when viewing thread) */}
          <div
            className={cn(
              'flex-1 flex flex-col',
              !isMobileViewingThread ? 'hidden md:flex' : 'flex'
            )}
          >
            <MessageThread
              conversation={pendingRecipient ? {
                id: 'pending',
                other_user: pendingRecipient,
                unread_count: 0,
                created_at: new Date().toISOString(),
              } : currentConversation}
              messages={messages}
              currentUserId={user?.id || ''}
              isLoading={isLoading && !!currentConversation}
              isSending={isSending}
              onSendMessage={handleSendMessage}
              onBack={handleBackToList}
              showBackButton={isMobileViewingThread}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
