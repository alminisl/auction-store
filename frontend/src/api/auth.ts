import api, { setAccessToken } from './client';
import {
  APIResponse,
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  VerifyEmailRequest,
  User,
} from '../types';

export const authApi = {
  async register(data: RegisterRequest): Promise<APIResponse<AuthResponse>> {
    const response = await api.post<APIResponse<AuthResponse>>('/auth/register', data);
    if (response.data.success && response.data.data) {
      setAccessToken(response.data.data.access_token);
    }
    return response.data;
  },

  async login(data: LoginRequest): Promise<APIResponse<AuthResponse>> {
    const response = await api.post<APIResponse<AuthResponse>>('/auth/login', data);
    if (response.data.success && response.data.data) {
      setAccessToken(response.data.data.access_token);
    }
    return response.data;
  },

  async logout(): Promise<APIResponse<void>> {
    const response = await api.post<APIResponse<void>>('/auth/logout');
    setAccessToken(null);
    return response.data;
  },

  async getMe(): Promise<APIResponse<User>> {
    const response = await api.get<APIResponse<User>>('/users/me');
    return response.data;
  },

  async forgotPassword(data: ForgotPasswordRequest): Promise<APIResponse<void>> {
    const response = await api.post<APIResponse<void>>('/auth/forgot-password', data);
    return response.data;
  },

  async resetPassword(data: ResetPasswordRequest): Promise<APIResponse<void>> {
    const response = await api.post<APIResponse<void>>('/auth/reset-password', data);
    return response.data;
  },

  async verifyEmail(data: VerifyEmailRequest): Promise<APIResponse<void>> {
    const response = await api.post<APIResponse<void>>('/auth/verify-email', data);
    return response.data;
  },

  async resendVerification(): Promise<APIResponse<void>> {
    const response = await api.post<APIResponse<void>>('/auth/resend-verification');
    return response.data;
  },

  getGoogleAuthUrl(): string {
    return '/api/auth/google';
  },
};

export default authApi;
