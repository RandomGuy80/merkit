'use client'

import { useAuthStore } from '@/lib/store/auth'
import { authApi } from '@/lib/api/auth'
import { useRouter } from 'next/navigation'

export function useAuth() {
  const { user, setAuth, clearAuth } = useAuthStore()
  const router = useRouter()

  const login = async (email: string, password: string) => {
    const resp = await authApi.login({ email, password })
    setAuth(resp.user, resp.access_token)
    return resp
  }

  const register = async (
    email: string,
    password: string,
    name: string,
    role: 'buyer' | 'seller'
  ) => {
    const resp = await authApi.register({ email, password, name, role })
    setAuth(resp.user, resp.access_token)
    return resp
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } finally {
      clearAuth()
      router.push('/login')
    }
  }

  return { user, login, register, logout, isAuthenticated: !!user }
}
