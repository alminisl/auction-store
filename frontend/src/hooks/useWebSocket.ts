import { useEffect, useRef, useCallback, useState } from 'react';
import { getAccessToken } from '../api/client';

interface WebSocketMessage {
  type: string;
  data: unknown;
}

interface UseWebSocketOptions {
  onMessage?: (message: WebSocketMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  reconnectAttempts?: number;
  reconnectInterval?: number;
}

export function useWebSocket(auctionId: string | null, options: UseWebSocketOptions = {}) {
  const {
    onMessage,
    onConnect,
    onDisconnect,
    reconnectAttempts = 5,
    reconnectInterval = 3000,
  } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectCountRef = useRef(0);
  const [isConnected, setIsConnected] = useState(false);

  const connect = useCallback(() => {
    if (!auctionId) return;

    const token = getAccessToken();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/auctions/${auctionId}${token ? `?token=${token}` : ''}`;

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setIsConnected(true);
      reconnectCountRef.current = 0;
      onConnect?.();
    };

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        onMessage?.(message);
      } catch {
        console.error('Failed to parse WebSocket message');
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      onDisconnect?.();

      // Attempt to reconnect
      if (reconnectCountRef.current < reconnectAttempts) {
        reconnectCountRef.current += 1;
        setTimeout(connect, reconnectInterval);
      }
    };

    ws.onerror = () => {
      ws.close();
    };

    wsRef.current = ws;
  }, [auctionId, onMessage, onConnect, onDisconnect, reconnectAttempts, reconnectInterval]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      reconnectCountRef.current = reconnectAttempts; // Prevent reconnection
      wsRef.current.close();
      wsRef.current = null;
    }
  }, [reconnectAttempts]);

  const sendMessage = useCallback((message: WebSocketMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  return {
    isConnected,
    sendMessage,
    disconnect,
    reconnect: connect,
  };
}

export default useWebSocket;
