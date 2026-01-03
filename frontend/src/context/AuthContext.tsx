import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import type { ReactNode } from 'react'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { StockCheckerService } from '../gen/stockchecker/v1/service_pb.js'
import type { User } from '../gen/stockchecker/v1/service_pb.js'

interface AuthContextType {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: () => void
  logout: () => void
  refetchUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// Custom fetch that includes credentials
const fetchWithCredentials: typeof fetch = (input, init) => {
  return fetch(input, { ...init, credentials: 'include' })
}

// Create transport with credentials to send cookies
const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  fetch: fetchWithCredentials,
})

const client = createClient(StockCheckerService, transport)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const fetchUser = useCallback(async () => {
    try {
      const response = await client.getCurrentUser({})
      setUser(response.user ?? null)
    } catch {
      // Not authenticated or auth not enabled
      setUser(null)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchUser()
  }, [fetchUser])

  // Check for auth error in URL params (from OAuth callback)
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const error = params.get('error')
    if (error === 'not_allowed') {
      // Clear the URL params
      window.history.replaceState({}, '', window.location.pathname)
      alert('Your email is not on the allowed list. Contact the administrator for access.')
    }
  }, [])

  const login = () => {
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    window.location.href = `${apiUrl}/auth/login`
  }

  const logout = () => {
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    window.location.href = `${apiUrl}/auth/logout`
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: user !== null,
        login,
        logout,
        refetchUser: fetchUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
