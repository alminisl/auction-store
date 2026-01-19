import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User } from '../types';
import { authApi } from '../api/auth';
import { setAccessToken } from '../api/client';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchUser: () => Promise<void>;
  clearError: () => void;
  setUser: (user: User | null) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (email: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authApi.login({ email, password });
          if (response.success && response.data) {
            set({
              user: response.data.user,
              isAuthenticated: true,
              isLoading: false,
            });
          } else {
            set({
              error: response.error?.message || 'Login failed',
              isLoading: false,
            });
          }
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Login failed';
          set({ error: message, isLoading: false });
        }
      },

      register: async (email: string, username: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authApi.register({ email, username, password });
          if (response.success && response.data) {
            set({
              user: response.data.user,
              isAuthenticated: true,
              isLoading: false,
            });
          } else {
            set({
              error: response.error?.message || 'Registration failed',
              isLoading: false,
            });
          }
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Registration failed';
          set({ error: message, isLoading: false });
        }
      },

      logout: async () => {
        try {
          await authApi.logout();
        } catch {
          // Ignore logout errors
        } finally {
          setAccessToken(null);
          set({ user: null, isAuthenticated: false });
        }
      },

      fetchUser: async () => {
        if (!get().isAuthenticated) return;

        set({ isLoading: true });
        try {
          const response = await authApi.getMe();
          if (response.success && response.data) {
            set({ user: response.data, isLoading: false });
          } else {
            set({ isAuthenticated: false, user: null, isLoading: false });
          }
        } catch {
          set({ isAuthenticated: false, user: null, isLoading: false });
        }
      },

      clearError: () => set({ error: null }),

      setUser: (user: User | null) => set({ user, isAuthenticated: !!user }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

export default useAuthStore;
