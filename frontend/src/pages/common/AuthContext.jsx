import React, { createContext, useContext, useState, useEffect } from 'react';
import { getUserInfo } from '/src/pages/common/auth_apis';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [session, setSession] = useState(() => {
    const stored = localStorage.getItem('session');
    return stored ? JSON.parse(stored) : null;
  });

  const login = (token, email, name) => {
    const s = { token, email, name };
    setSession(s);
    localStorage.setItem('session', JSON.stringify(s));
  };

  const logout = () => {
    setSession(null);
    localStorage.removeItem('session');
  };

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
    <AuthContext.Provider value={{ isAuthenticated: !!session, session, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
