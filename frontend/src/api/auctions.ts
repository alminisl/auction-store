import api from './client';
import {
  APIResponse,
  Auction,
  AuctionListParams,
  Category,
  CreateAuctionRequest,
  UpdateAuctionRequest,
  PaginatedResponse,
} from '../types';

export const auctionsApi = {
  async list(params?: AuctionListParams): Promise<APIResponse<Auction[]>> {
    const response = await api.get<APIResponse<Auction[]>>('/auctions', { params });
    return response.data;
  },

  async getById(id: string): Promise<APIResponse<Auction>> {
    const response = await api.get<APIResponse<Auction>>(`/auctions/${id}`);
    return response.data;
  },

  async create(data: CreateAuctionRequest): Promise<APIResponse<Auction>> {
    const response = await api.post<APIResponse<Auction>>('/auctions', data);
    return response.data;
  },

  async update(id: string, data: UpdateAuctionRequest): Promise<APIResponse<Auction>> {
    const response = await api.put<APIResponse<Auction>>(`/auctions/${id}`, data);
    return response.data;
  },

  async delete(id: string): Promise<APIResponse<void>> {
    const response = await api.delete<APIResponse<void>>(`/auctions/${id}`);
    return response.data;
  },

  async publish(id: string): Promise<APIResponse<Auction>> {
    const response = await api.post<APIResponse<Auction>>(`/auctions/${id}/publish`);
    return response.data;
  },

  async uploadImage(id: string, file: File): Promise<APIResponse<{ id: string; url: string; position: number }>> {
    const formData = new FormData();
    formData.append('image', file);

    const response = await api.post<APIResponse<{ id: string; url: string; position: number }>>(`/auctions/${id}/images`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  async deleteImage(auctionId: string, imageId: string): Promise<APIResponse<void>> {
    const response = await api.delete<APIResponse<void>>(`/auctions/${auctionId}/images/${imageId}`);
    return response.data;
  },

  async getCategories(): Promise<APIResponse<Category[]>> {
    const response = await api.get<APIResponse<Category[]>>('/categories');
    return response.data;
  },

  async getCategoryBySlug(slug: string): Promise<APIResponse<Category>> {
    const response = await api.get<APIResponse<Category>>(`/categories/${slug}`);
    return response.data;
  },

  async getMyAuctions(params?: { page?: number; limit?: number; status?: string }): Promise<APIResponse<Auction[]>> {
    const response = await api.get<APIResponse<Auction[]>>('/users/me/auctions', { params });
    return response.data;
  },

  async getUserAuctions(userId: string, params?: { page?: number; limit?: number }): Promise<APIResponse<Auction[]>> {
    const response = await api.get<APIResponse<Auction[]>>(`/users/${userId}/auctions`, { params });
    return response.data;
  },
};

export default auctionsApi;
