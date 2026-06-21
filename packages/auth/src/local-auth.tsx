"use client";

import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

// ---------- Types ----------

interface LocalUser {
  createdAt: string;
  email: string;
  id: string;
  updatedAt: string;
  username: string;
}

interface AuthResult {
  token: string;
  user: LocalUser;
}

interface LocalAuthState {
  isLoaded: boolean;
  isSignedIn: boolean;
  token: string | null;
  user: LocalUser | null;
}

interface LocalAuthContextValue extends LocalAuthState {
  /** Base URL for the auth API, e.g. http://localhost:3002/api/v1 */
  apiUrl: string;
  getToken: () => Promise<string | null>;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => void;
  signUp: (username: string, email: string, password: string) => Promise<void>;
}

// ---------- Context ----------

const LocalAuthContext = createContext<LocalAuthContextValue | null>(null);

// ---------- Storage ----------

const TOKEN_KEY = "starterpack_auth_token";
const USER_KEY = "starterpack_auth_user";

function persistAuth(token: string, user: LocalUser) {
  try {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  } catch {
    // SSR or storage full — ignore
  }
}

function clearAuth() {
  try {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  } catch {
    // ignore
  }
}

function loadAuth(): { token: string | null; user: LocalUser | null } {
  try {
    const token = localStorage.getItem(TOKEN_KEY);
    const raw = localStorage.getItem(USER_KEY);
    const user = raw ? (JSON.parse(raw) as LocalUser) : null;
    return { token, user };
  } catch {
    return { token: null, user: null };
  }
}

// ---------- Provider ----------

interface LocalAuthProviderProps {
  apiUrl: string;
  children: ReactNode;
}

/**
 * LocalAuthProvider manages local username/password auth state. It stores the
 * JWT in localStorage and validates it on mount by calling /auth/me.
 */
export function LocalAuthProvider({
  apiUrl,
  children,
}: LocalAuthProviderProps) {
  const [state, setState] = useState<LocalAuthState>({
    isLoaded: false,
    isSignedIn: false,
    token: null,
    user: null,
  });

  // On mount, validate stored token.
  useEffect(() => {
    const stored = loadAuth();
    if (!stored.token) {
      setState({ isLoaded: true, isSignedIn: false, token: null, user: null });
      return;
    }
    // Validate token by calling /auth/me.
    fetch(`${apiUrl}/auth/me`, {
      headers: { Authorization: `Bearer ${stored.token}` },
    })
      .then(async (res) => {
        if (res.ok) {
          const user = (await res.json()) as LocalUser;
          setState({
            isLoaded: true,
            isSignedIn: true,
            token: stored.token,
            user,
          });
        } else {
          clearAuth();
          setState({
            isLoaded: true,
            isSignedIn: false,
            token: null,
            user: null,
          });
        }
      })
      .catch(() => {
        // Network error — keep stored state as optimistic fallback.
        if (stored.user) {
          setState({
            isLoaded: true,
            isSignedIn: true,
            token: stored.token,
            user: stored.user,
          });
        } else {
          clearAuth();
          setState({
            isLoaded: true,
            isSignedIn: false,
            token: null,
            user: null,
          });
        }
      });
  }, [apiUrl]);

  const signIn = useCallback(
    async (email: string, password: string) => {
      const res = await fetch(`${apiUrl}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as {
          error?: string;
        };
        throw new Error(body.error ?? "Login failed");
      }
      const data = (await res.json()) as AuthResult;
      persistAuth(data.token, data.user);
      setState({
        isLoaded: true,
        isSignedIn: true,
        token: data.token,
        user: data.user,
      });
    },
    [apiUrl]
  );

  const signUp = useCallback(
    async (username: string, email: string, password: string) => {
      const res = await fetch(`${apiUrl}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, email, password }),
      });
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as {
          error?: string;
        };
        throw new Error(body.error ?? "Registration failed");
      }
      const data = (await res.json()) as AuthResult;
      persistAuth(data.token, data.user);
      setState({
        isLoaded: true,
        isSignedIn: true,
        token: data.token,
        user: data.user,
      });
    },
    [apiUrl]
  );

  const signOut = useCallback(() => {
    clearAuth();
    setState({
      isLoaded: true,
      isSignedIn: false,
      token: null,
      user: null,
    });
  }, []);

  const getToken = useCallback(async () => {
    return state.token;
  }, [state.token]);

  const value = useMemo<LocalAuthContextValue>(
    () => ({
      ...state,
      apiUrl,
      signIn,
      signUp,
      signOut,
      getToken,
    }),
    [state, apiUrl, signIn, signUp, signOut, getToken]
  );

  return (
    <LocalAuthContext.Provider value={value}>
      {children}
    </LocalAuthContext.Provider>
  );
}

// ---------- Hooks ----------

/**
 * useLocalAuth returns the local auth context. Must be used within
 * LocalAuthProvider.
 */
export function useLocalAuth(): LocalAuthContextValue {
  const ctx = useContext(LocalAuthContext);
  if (!ctx) {
    throw new Error("useLocalAuth must be used within <LocalAuthProvider>");
  }
  return ctx;
}

export type { LocalUser, LocalAuthContextValue };
