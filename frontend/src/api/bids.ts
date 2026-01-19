import api from './client';
import { APIResponse, Bid, BidResponse, PlaceBidRequest, PaginatedResponse } from '../types';

export const bidsApi = {
  async placeBid(auctionId: string, data: PlaceBidRequest): Promise<APIResponse<BidResponse>> {
    const response = await api.post<APIResponse<BidResponse>>(`/auctions/${auctionId}/bids`, data);
    return response.data;
  },

  async getBidsByAuction(
    auctionId: string,
    params?: { page?: number; limit?: number }
  ): Promise<APIResponse<PaginatedResponse<Bid>>> {
    const response = await api.get<APIResponse<PaginatedResponse<Bid>>>(`/auctions/${auctionId}/bids`, { params });
    return response.data;
  },

  async getMyBids(params?: { page?: number; limit?: number }): Promise<APIResponse<Bid[]>> {
    const response = await api.get<APIResponse<Bid[]>>('/users/me/bids', { params });
    return response.data;
  },

  async buyNow(auctionId: string): Promise<APIResponse<BidResponse>> {
    const response = await api.post<APIResponse<BidResponse>>(`/auctions/${auctionId}/buy-now`);
    return response.data;
  },
};

export default bidsApi;
