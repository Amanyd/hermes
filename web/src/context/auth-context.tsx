'use client'

import { login as apiLogin, register as apiRegister } from '@/lib/api'
import type { LoginRequest, RegisterRequest, User } from '@/types/auth'
import { useRouter } from 'next/navigation'
import { createContext, useCallback, useContext, useEffect, useState } from 'react'

interface AuthContextValue {
  user: User | null
  token: string | null
  isLoading: boolean
  login: (data: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  // Rehydrate from localStorage on mount
  useEffect(() => {
    const storedToken = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user')
    if (storedToken && storedUser) {
      setToken(storedToken)
      setUser(JSON.parse(storedUser))
    }
    setIsLoading(false)
  }, [])

  const persist = useCallback((token: string, user: User) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    // Also write to cookie so middleware can read it server-side
    document.cookie = `token=${token}; path=/; max-age=${60 * 60 * 24 * 7}; SameSite=Lax`
    setToken(token)
    setUser(user)
  }, [])

  const login = useCallback(
    async (data: LoginRequest) => {
      const res = await apiLogin(data)
      persist(res.token, res.user)
      router.push('/dashboard/relays')
    },
    [persist, router],
  )

  const register = useCallback(
    async (data: RegisterRequest) => {
      const res = await apiRegister(data)
      persist(res.token, res.user)
      router.push('/dashboard/relays')
    },
    [persist, router],
  )

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    // Clear cookie
    document.cookie = 'token=; path=/; max-age=0'
    setToken(null)
    setUser(null)
    router.push('/login')
  }, [router])

  return (
    <AuthContext.Provider value={{ user, token, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>')
  return ctx
}
