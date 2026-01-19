export interface User {
  id: string;
  email: string;
  username: string;
  avatar_url?: string;
  bio?: string;
  phone?: string;
  address?: string;
  role: 'user' | 'admin';
  email_verified: boolean;
  created_at: string;
  updated_at: string;
  rating_summary?: UserRatingSummary;
}

export interface UserRatingSummary {
  user_id: string;
  average_rating: number;
  total_ratings: number;
  positive_ratings: number;
  neutral_ratings: number;
  negative_ratings: number;
}

export interface PublicUser {
  id: string;
  username: string;
  avatar_url?: string;
  bio?: string;
  created_at: string;
  rating_summary?: UserRatingSummary;
}

export interface UpdateProfileRequest {
  username?: string;
  bio?: string;
  phone?: string;
  address?: string;
}
