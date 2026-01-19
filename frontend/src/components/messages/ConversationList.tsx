import { User } from 'lucide-react';
import { cn } from '../../utils/cn';
import { Conversation } from '../../types';
import { formatDistanceToNow } from 'date-fns';

interface ConversationListProps {
  conversations: Conversation[];
  selectedId?: string;
  onSelect: (conversation: Conversation) => void;
}

interface ConversationItemProps {
  conversation: Conversation;
  isSelected: boolean;
  onClick: () => void;
}

function ConversationItem({ conversation, isSelected, onClick }: ConversationItemProps) {
  const otherUser = conversation.other_user;
  const lastMessage = conversation.last_message;
  const hasUnread = conversation.unread_count > 0;

  const timeAgo = conversation.last_message_at
    ? formatDistanceToNow(new Date(conversation.last_message_at), { addSuffix: true })
    : '';

  // Truncate message content
  const truncatedContent = lastMessage?.content
    ? lastMessage.content.length > 50
      ? lastMessage.content.slice(0, 50) + '...'
      : lastMessage.content
    : 'No messages yet';

  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full p-4 flex items-start gap-3 hover:bg-muted/50 transition-colors text-left',
        isSelected && 'bg-muted',
        hasUnread && 'bg-primary/5'
      )}
    >
      {/* Avatar */}
      <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center overflow-hidden shrink-0">
        {otherUser.avatar_url ? (
          <img
            src={otherUser.avatar_url}
            alt={otherUser.username}
            className="w-full h-full object-cover"
          />
        ) : (
          <User className="h-6 w-6 text-muted-foreground" />
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <span className={cn('font-medium truncate', hasUnread && 'text-foreground')}>
            {otherUser.username}
          </span>
          <span className="text-xs text-muted-foreground shrink-0">{timeAgo}</span>
        </div>
        <div className="flex items-center justify-between gap-2 mt-1">
          <p
            className={cn(
              'text-sm truncate',
              hasUnread ? 'text-foreground font-medium' : 'text-muted-foreground'
            )}
          >
            {truncatedContent}
          </p>
          {hasUnread && (
            <span className="bg-primary text-primary-foreground text-xs font-medium px-2 py-0.5 rounded-full shrink-0">
              {conversation.unread_count}
            </span>
          )}
        </div>
      </div>
    </button>
  );
}

export function ConversationList({
  conversations,
  selectedId,
  onSelect,
}: ConversationListProps) {
  if (conversations.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center p-8 text-center text-muted-foreground">
        <div>
          <p>No conversations yet</p>
          <p className="text-sm mt-1">Start a conversation from a user's profile</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto divide-y divide-border">
      {conversations.map((conversation) => (
        <ConversationItem
          key={conversation.id}
          conversation={conversation}
          isSelected={selectedId === conversation.id}
          onClick={() => onSelect(conversation)}
        />
      ))}
    </div>
  );
}

export default ConversationList;
