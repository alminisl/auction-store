export interface APIResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: APIError;
  meta?: APIMeta;
}

export interface APIError {
  code: string;
  message: string;
  details?: Record<string, string>;
}

export interface APIMeta {
  page?: number;
  limit?: number;
  total?: number;
  total_pages?: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  meta: APIMeta;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface AuthResponse {
  user: import('./user').User;
  access_token: string;
  expires_in: number;
}

export interface RefreshResponse {
  access_token: string;
  expires_in: number;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  password: string;
}

export interface VerifyEmailRequest {
  token: string;
}
