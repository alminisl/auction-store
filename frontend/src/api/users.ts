import api from './client';
import {
  APIResponse,
  User,
  PublicUser,
  UpdateProfileRequest,
  Rating,
  Notification,
  WatchlistItem,
  PaginatedResponse,
  Auction,
} from '../types';

export const usersApi = {
  async getMe(): Promise<APIResponse<User>> {
    const response = await api.get<APIResponse<User>>('/users/me');
    return response.data;
  },

  async updateMe(data: UpdateProfileRequest): Promise<APIResponse<User>> {
    const response = await api.put<APIResponse<User>>('/users/me', data);
    return response.data;
  },

  async uploadAvatar(file: File): Promise<APIResponse<User>> {
    const formData = new FormData();
    formData.append('avatar', file);

    const response = await api.post<APIResponse<User>>('/users/me/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  async getUser(id: string): Promise<APIResponse<PublicUser>> {
    const response = await api.get<APIResponse<PublicUser>>(`/users/${id}`);
    return response.data;
  },

  async getUserRatings(
    userId: string,
    params?: { page?: number; limit?: number; type?: string }
  ): Promise<APIResponse<PaginatedResponse<Rating>>> {
    const response = await api.get<APIResponse<PaginatedResponse<Rating>>>(`/users/${userId}/ratings`, { params });
    return response.data;
  },

  // Notifications
  async getNotifications(params?: { page?: number; limit?: number }): Promise<APIResponse<PaginatedResponse<Notification>>> {
    const response = await api.get<APIResponse<PaginatedResponse<Notification>>>('/notifications', { params });
    return response.data;
  },

  async markNotificationRead(id: string): Promise<APIResponse<void>> {
    const response = await api.put<APIResponse<void>>(`/notifications/${id}/read`);
    return response.data;
  },

  async markAllNotificationsRead(): Promise<APIResponse<void>> {
    const response = await api.put<APIResponse<void>>('/notifications/read-all');
    return response.data;
  },

  // Watchlist
  async getWatchlist(params?: { page?: number; limit?: number }): Promise<APIResponse<WatchlistItem[]>> {
    const response = await api.get<APIResponse<WatchlistItem[]>>('/watchlist', { params });
    return response.data;
  },

  async addToWatchlist(auctionId: string): Promise<APIResponse<WatchlistItem>> {
    const response = await api.post<APIResponse<WatchlistItem>>(`/watchlist/${auctionId}`);
    return response.data;
  },

  async removeFromWatchlist(auctionId: string): Promise<APIResponse<void>> {
    const response = await api.delete<APIResponse<void>>(`/watchlist/${auctionId}`);
    return response.data;
  },

  // Ratings
  async createRating(
    auctionId: string,
    data: { rating: number; comment?: string; type: 'buyer' | 'seller' }
  ): Promise<APIResponse<Rating>> {
    const response = await api.post<APIResponse<Rating>>(`/auctions/${auctionId}/ratings`, data);
    return response.data;
  },

  // Won auctions
  async getWonAuctions(params?: { page?: number; limit?: number }): Promise<APIResponse<PaginatedResponse<Auction>>> {
    const response = await api.get<APIResponse<PaginatedResponse<Auction>>>('/users/me/won', { params });
    return response.data;
  },
};

export default usersApi;
