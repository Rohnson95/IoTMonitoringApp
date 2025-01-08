// src/context/AuthContext.tsx

import React, { createContext, useState, useEffect, ReactNode } from 'react';
import { jwtDecode } from 'jwt-decode';

interface AuthContextType {
    token: string | null;
    setToken: (token: string | null) => void;
    userId: number | null;
    logout: () => void;
}

interface TokenPayload {
    user_id: number;
    exp: number;
    iat: number;
    iss: string;
}

export const AuthContext = createContext<AuthContextType>({
    token: null,
    setToken: () => {},
    userId: null,
    logout: () => {},
});

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
    const [userId, setUserId] = useState<number | null>(null);

    useEffect(() => {
        if (token) {
            localStorage.setItem('token', token);
            try {
                const decoded = jwtDecode<TokenPayload>(token);
                setUserId(decoded.user_id);
            } catch (e) {
                console.error("Failed to decode token", e);
                setUserId(null);
            }
        } else {
            localStorage.removeItem('token');
            setUserId(null);
        }
    }, [token]);

    const logout = () => {
        setToken(null);
    }

    return (
        <AuthContext.Provider value={{ token, setToken, userId, logout }}>
            {children}
        </AuthContext.Provider>
    );
};
