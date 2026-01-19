import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { APIResponse, RefreshResponse } from '../types';

const BASE_URL = '/api';

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // For refresh token cookies
});

// Token management
let accessToken: string | null = null;
let refreshPromise: Promise<string | null> | null = null;

export const setAccessToken = (token: string | null) => {
  accessToken = token;
};

export const getAccessToken = () => accessToken;

// Request interceptor to add auth header
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (accessToken && config.headers) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor for token refresh
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<APIResponse>) => {
    const originalRequest = error.config;

    // If unauthorized and not a refresh request
    if (
      error.response?.status === 401 &&
      originalRequest &&
      !originalRequest.url?.includes('/auth/refresh') &&
      !originalRequest.url?.includes('/auth/login')
    ) {
      // Try to refresh the token
      if (!refreshPromise) {
        refreshPromise = refreshAccessToken();
      }

      try {
        const newToken = await refreshPromise;
        refreshPromise = null;

        if (newToken && originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return api(originalRequest);
        }
      } catch {
        refreshPromise = null;
        // Refresh failed, clear token and redirect to login
        setAccessToken(null);
        window.location.href = '/login';
      }
    }

    return Promise.reject(error);
  }
);

// Refresh access token
async function refreshAccessToken(): Promise<string | null> {
  try {
    const response = await axios.post<APIResponse<RefreshResponse>>(
      `${BASE_URL}/auth/refresh`,
      {},
      { withCredentials: true }
    );

    if (response.data.success && response.data.data) {
      const { access_token } = response.data.data;
      setAccessToken(access_token);
      return access_token;
    }
    return null;
  } catch {
    return null;
  }
}

export default api;
