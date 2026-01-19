import { Auction } from './auction';
import { PublicUser } from './user';

export interface Bid {
  id: string;
  auction_id: string;
  auction?: Auction;
  bidder_id: string;
  bidder?: PublicUser;
  amount: string;
  is_auto_bid: boolean;
  max_auto_bid?: string;
  created_at: string;
}

export interface PlaceBidRequest {
  amount: string;
  is_auto_bid?: boolean;
  max_auto_bid?: string;
}

export interface BidResponse {
  bid: Bid;
  auction: Auction;
}
