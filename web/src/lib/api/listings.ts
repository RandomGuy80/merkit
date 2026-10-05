import api from './client'
import type { Listing, ListingsPage, ListingsFilter } from '@/types'

export const listingsApi = {
  search: (f: ListingsFilter = {}) =>
    api.get<ListingsPage>('/listings', { params: f }).then((r) => r.data),

  getById: (id: string) =>
    api.get<Listing>(`/listings/${id}`).then((r) => r.data),

  create: (data: {
    title: string
    description: string
    price: number
    currency?: string
    category_id?: number
    location?: string
    tags?: string[]
  }) => api.post<Listing>('/listings', data).then((r) => r.data),

  update: (id: string, data: Partial<{
    title: string
    description: string
    price: number
    category_id: number
    location: string
    tags: string[]
    status: string
  }>) => api.put<Listing>(`/listings/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    api.delete(`/listings/${id}`),

  addImage: (id: string, file: File) => {
    const form = new FormData()
    form.append('image', file)
    return api
      .post<{ image_url: string }>(`/listings/${id}/images`, form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((r) => r.data)
  },

  removeImage: (id: string, index: number) =>
    api.delete(`/listings/${id}/images/${index}`),
}
