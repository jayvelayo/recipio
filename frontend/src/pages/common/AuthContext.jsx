import React, { createContext, useContext, useState, useEffect } from 'react';
import { getUserInfo } from '/src/pages/common/auth_apis';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [session, setSession] = useState(() => {
    const stored = localStorage.getItem('session');
    return stored ? JSON.parse(stored) : null;
  });

  const [sessionExpired, setSessionExpired] = useState(false);

  const login = (token, email, name) => {
    const s = { token, email, name };
    setSession(s);
    setSessionExpired(false);
    localStorage.setItem('session', JSON.stringify(s));
  };

  const logout = () => {
    setSession(null);
    localStorage.removeItem('session');
  };

  // Sign out and prompt for re-authentication when the server rejects a stored
  // session (401). Login/register are excluded: a 401 there means bad credentials.
  useEffect(() => {
    const originalFetch = window.fetch;
    window.fetch = async (input, init) => {
      const res = await originalFetch(input, init);
      const url = typeof input === 'string' ? input : input.url;
      if (res.status === 401 && localStorage.getItem('session') && !url.startsWith('/auth/login') && !url.startsWith('/auth/register')) {
        logout();
        setSessionExpired(true);
      }
      return res;
    };
    return () => { window.fetch = originalFetch; };
  }, []);

  // Handle the Google OAuth redirect: backend sends /#google_token=<session_token>
  useEffect(() => {
    const match = window.location.hash.match(/google_token=([^&]+)/);
    if (!match) return;
    const token = match[1];
    window.history.replaceState(null, '', window.location.pathname + window.location.search);
    getUserInfo(token)
      .then(({ name, email }) => login(token, email, name))
      .catch(() => {});
  }, []);

  return (
    <AuthContext.Provider value={{ isAuthenticated: !!session, session, sessionExpired, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
