'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState, useEffect, ReactNode } from 'react'
import { ToastProvider } from '@/components/ui/Toast'
import { useAuthStore } from '@/lib/store/auth'
import { usersApi } from '@/lib/api/users'

function AuthHydrator() {
  const { setAuth, accessToken, setHydrated } = useAuthStore()

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token || accessToken) {
      setHydrated()
      return
    }
    usersApi.getMe()
      .then((user) => setAuth(user, token))
      .catch(() => {})
      .finally(() => setHydrated())
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return null
}

export function Providers({ children }: { children: ReactNode }) {
  const [qc] = useState(() => new QueryClient({
    defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
  }))

  return (
    <QueryClientProvider client={qc}>
      <ToastProvider>
        <AuthHydrator />
        {children}
      </ToastProvider>
    </QueryClientProvider>
  )
}
