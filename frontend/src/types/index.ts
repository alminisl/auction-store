export * from './user';
export * from './auction';
export * from './bid';
export * from './api';

export interface Notification {
  id: string;
  user_id: string;
  type: 'outbid' | 'auction_won' | 'auction_lost' | 'auction_ending' | 'new_bid' | 'watchlist_ending';
  title: string;
  message?: string;
  auction_id?: string;
  is_read: boolean;
  created_at: string;
}

export interface Rating {
  id: string;
  auction_id: string;
  rater_id: string;
  rated_user_id: string;
  rating: number;
  comment?: string;
  type: 'buyer' | 'seller';
  created_at: string;
}

export interface WatchlistItem {
  id: string;
  user_id: string;
  auction_id: string;
  auction?: import('./auction').Auction;
  created_at: string;
}
