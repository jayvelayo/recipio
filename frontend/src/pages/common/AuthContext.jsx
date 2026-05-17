import React, { createContext, useContext, useState } from 'react';

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

  return (
    <AuthContext.Provider value={{ isAuthenticated: !!session, session, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
