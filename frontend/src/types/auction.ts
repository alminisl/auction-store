import { PublicUser } from './user';

export type AuctionStatus = 'draft' | 'active' | 'completed' | 'cancelled' | 'unsold';
// Card conditions (for trading cards)
export type CardCondition = 'mint' | 'near_mint' | 'excellent' | 'good' | 'played';
// General conditions (for other items)
export type GeneralCondition = 'new' | 'like_new' | 'very_good' | 'good' | 'acceptable';
export type AuctionCondition = CardCondition | GeneralCondition;
export type AuctionCurrency = 'USD' | 'EUR' | 'BAM';

// Trading card category slugs (used to determine which conditions to show)
export const TRADING_CARD_CATEGORY_SLUGS = ['trading-cards', 'tcg', 'pokemon', 'magic-the-gathering', 'yugioh', 'one-piece', 'sports-cards'];

export interface AuctionImage {
  id: string;
  auction_id: string;
  url: string;
  position: number;
  created_at: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  parent_id?: string;
  description?: string;
  image_url?: string;
  auction_count?: number;
}

export interface Auction {
  id: string;
  seller_id: string;
  seller?: PublicUser;
  category_id?: string;
  category?: Category;
  title: string;
  description?: string;
  condition?: AuctionCondition;
  currency: AuctionCurrency;
  starting_price: string;
  reserve_price?: string;
  buy_now_price?: string;
  current_price: string;
  bid_increment: string;
  start_time: string;
  end_time: string;
  status: AuctionStatus;
  winner_id?: string;
  winner?: PublicUser;
  winning_bid_id?: string;
  views_count: number;
  bid_count: number;
  images: AuctionImage[];
  is_watched?: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateAuctionRequest {
  title: string;
  description?: string;
  category_id?: string;
  condition?: AuctionCondition;
  currency?: AuctionCurrency;
  starting_price: string;
  reserve_price?: string;
  buy_now_price?: string;
  bid_increment?: string;
  start_time: string;
  end_time: string;
}

export interface UpdateAuctionRequest {
  title?: string;
  description?: string;
  category_id?: string;
  condition?: AuctionCondition;
  reserve_price?: string;
  buy_now_price?: string;
  bid_increment?: string;
  start_time?: string;
  end_time?: string;
}

export interface AuctionListParams {
  page?: number;
  limit?: number;
  status?: AuctionStatus;
  category_id?: string;
  seller_id?: string;
  search?: string;
  sort_by?: 'created_at' | 'end_time' | 'current_price' | 'bid_count';
  sort_order?: 'asc' | 'desc';
  min_price?: string;
  max_price?: string;
}
