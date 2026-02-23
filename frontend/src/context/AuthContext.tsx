import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { authApi, getToken, clearTokens, type User } from '../api/client';

interface AuthState {
  isLoggedIn: boolean;
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  googleLogin: (code: string, redirectUri: string) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchMe = useCallback(async () => {
    const token = getToken();
    if (!token) {
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      const u = await authApi.me();
      setUser(u);
    } catch {
      clearTokens();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchMe(); }, [fetchMe]);

  const login = useCallback(async (email: string, password: string) => {
    await authApi.login(email, password);
    await fetchMe();
  }, [fetchMe]);

  const register = useCallback(async (email: string, password: string) => {
    await authApi.register(email, password);
    await fetchMe();
  }, [fetchMe]);

  const googleLogin = useCallback(async (code: string, redirectUri: string) => {
    await authApi.googleAuth(code, redirectUri);
    await fetchMe();
  }, [fetchMe]);

  const logout = useCallback(() => {
    authApi.logout();
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{
      isLoggedIn: !!user,
      user,
      loading,
      login,
      register,
      googleLogin,
      logout,
      refreshUser: fetchMe,
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
