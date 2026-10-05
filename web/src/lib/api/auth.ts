import api from './client'
import type { AuthResponse } from '@/types'

export const authApi = {
  register: (data: { email: string; password: string; name: string; role: string }) =>
    api.post<AuthResponse>('/auth/register', data).then((r) => r.data),

  login: (data: { email: string; password: string }) =>
    api.post<AuthResponse>('/auth/login', data).then((r) => r.data),

  refresh: () =>
    api.post<AuthResponse>('/auth/refresh', {}).then((r) => r.data),

  logout: () =>
    api.post('/auth/logout', {}).then((r) => r.data),
}
