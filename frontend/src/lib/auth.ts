import { create } from 'zustand';
import { User } from '@/types';

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
  checkAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isLoading: true,
  setAuth: (user, token) => {
    localStorage.setItem('token', token);
    set({ user, token, isLoading: false });
  },
  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null, isLoading: false });
  },
  checkAuth: () => {
    const token = localStorage.getItem('token');
    if (!token) {
      set({ user: null, token: null, isLoading: false });
      return;
    }

    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const isExpired = payload.exp * 1000 < Date.now();
      
      if (isExpired) {
        localStorage.removeItem('token');
        set({ user: null, token: null, isLoading: false });
        return;
      }

      set({
        token,
        user: {
          id: payload.user_id,
          email: payload.email,
          role: payload.role,
          created_at: '',
        },
        isLoading: false,
      });
    } catch {
      localStorage.removeItem('token');
      set({ user: null, token: null, isLoading: false });
    }
  },
}));
