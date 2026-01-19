import { cn } from '../../utils/cn';
import { MessageWithSender } from '../../types';
import { format } from 'date-fns';

interface MessageBubbleProps {
  message: MessageWithSender;
  isOwn: boolean;
}

export function MessageBubble({ message, isOwn }: MessageBubbleProps) {
  const time = format(new Date(message.created_at), 'HH:mm');

  return (
    <div className={cn('flex', isOwn ? 'justify-end' : 'justify-start')}>
      <div
        className={cn(
          'max-w-[70%] rounded-lg px-4 py-2',
          isOwn
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted text-foreground'
        )}
      >
        <p className="whitespace-pre-wrap break-words">{message.content}</p>
        <p
          className={cn(
            'text-xs mt-1',
            isOwn ? 'text-primary-foreground/70' : 'text-muted-foreground'
          )}
        >
          {time}
        </p>
      </div>
    </div>
  );
}

export default MessageBubble;
