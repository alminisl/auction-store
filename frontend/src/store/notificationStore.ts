import { create } from 'zustand';
import { Notification } from '../types';
import { usersApi } from '../api/users';

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  isLoading: boolean;

  // Actions
  fetchNotifications: () => Promise<void>;
  markAsRead: (id: string) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  addNotification: (notification: Notification) => void;
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
  notifications: [],
  unreadCount: 0,
  isLoading: false,

  fetchNotifications: async () => {
    set({ isLoading: true });
    try {
      const response = await usersApi.getNotifications({ limit: 50 });
      if (response.success && response.data) {
        const notifications = response.data.items;
        const unreadCount = notifications.filter((n) => !n.is_read).length;
        set({ notifications, unreadCount, isLoading: false });
      }
    } catch {
      set({ isLoading: false });
    }
  },

  markAsRead: async (id: string) => {
    try {
      await usersApi.markNotificationRead(id);
      const notifications = get().notifications.map((n) =>
        n.id === id ? { ...n, is_read: true } : n
      );
      const unreadCount = notifications.filter((n) => !n.is_read).length;
      set({ notifications, unreadCount });
    } catch {
      // Ignore errors
    }
  },

  markAllAsRead: async () => {
    try {
      await usersApi.markAllNotificationsRead();
      const notifications = get().notifications.map((n) => ({ ...n, is_read: true }));
      set({ notifications, unreadCount: 0 });
    } catch {
      // Ignore errors
    }
  },

  addNotification: (notification: Notification) => {
    const notifications = [notification, ...get().notifications];
    const unreadCount = notification.is_read ? get().unreadCount : get().unreadCount + 1;
    set({ notifications, unreadCount });
  },
}));

export default useNotificationStore;
