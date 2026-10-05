import api from './client'
import type { User, Review, Category } from '@/types'

export const usersApi = {
  getMe: () =>
    api.get<User>('/users/me').then((r) => r.data),

  updateMe: (data: { name?: string; bio?: string }) =>
    api.put<User>('/users/me', data).then((r) => r.data),

  uploadAvatar: (file: File) => {
    const form = new FormData()
    form.append('avatar', file)
    return api
      .post<{ avatar_url: string }>('/users/me/avatar', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((r) => r.data)
  },

  getPublic: (id: string) =>
    api.get<User>(`/users/${id}`).then((r) => r.data),

  getReviews: (id: string) =>
    api.get<Review[]>(`/users/${id}/reviews`).then((r) => r.data),
}

export const categoriesApi = {
  list: () =>
    api.get<Category[]>('/categories').then((r) => r.data),
}
