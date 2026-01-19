import { useEffect, useRef, useCallback, useState } from 'react';
import { getAccessToken } from '../api/client';
import { WSMessagePayload } from '../types';
import { useMessageStore } from '../store/messageStore';

interface UseMessageWebSocketOptions {
  onMessage?: (payload: WSMessagePayload) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  reconnectAttempts?: number;
  reconnectInterval?: number;
  enabled?: boolean;
}

export function useMessageWebSocket(options: UseMessageWebSocketOptions = {}) {
  const {
    onMessage,
    onConnect,
    onDisconnect,
    reconnectAttempts = 5,
    reconnectInterval = 3000,
    enabled = true,
  } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectCountRef = useRef(0);
  const [isConnected, setIsConnected] = useState(false);
  const { addIncomingMessage } = useMessageStore();

  const connect = useCallback(() => {
    const token = getAccessToken();
    if (!token || !enabled) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/messages?token=${token}`;

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setIsConnected(true);
      reconnectCountRef.current = 0;
      onConnect?.();
    };

    ws.onmessage = (event) => {
      try {
        // Handle potential multiple messages separated by newlines
        const messages = event.data.split('\n').filter(Boolean);

        for (const msgStr of messages) {
          const payload: WSMessagePayload = JSON.parse(msgStr);

          // Handle new message
          if (payload.type === 'new_message' && payload.message && payload.conversation_id) {
            addIncomingMessage(payload.message, payload.conversation_id);
          }

          onMessage?.(payload);
        }
      } catch {
        console.error('Failed to parse WebSocket message');
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      onDisconnect?.();

      // Attempt to reconnect if enabled
      if (enabled && reconnectCountRef.current < reconnectAttempts) {
        reconnectCountRef.current += 1;
        setTimeout(connect, reconnectInterval);
      }
    };

    ws.onerror = () => {
      ws.close();
    };

    wsRef.current = ws;
  }, [enabled, onMessage, onConnect, onDisconnect, reconnectAttempts, reconnectInterval, addIncomingMessage]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      reconnectCountRef.current = reconnectAttempts; // Prevent reconnection
      wsRef.current.close();
      wsRef.current = null;
    }
  }, [reconnectAttempts]);

  useEffect(() => {
    if (enabled) {
      connect();
    }
    return () => disconnect();
  }, [connect, disconnect, enabled]);

  return {
    isConnected,
    disconnect,
    reconnect: connect,
  };
}

export default useMessageWebSocket;
