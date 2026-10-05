import api from './client'
import type { Order } from '@/types'

export const ordersApi = {
  create: (listing_id: string) =>
    api.post<Order>('/orders', { listing_id }).then((r) => r.data),

  list: () =>
    api.get<Order[]>('/orders').then((r) => r.data),

  getById: (id: string) =>
    api.get<Order>(`/orders/${id}`).then((r) => r.data),

  updateStatus: (id: string, status: string) =>
    api.patch<Order>(`/orders/${id}/status`, { status }).then((r) => r.data),

  checkout: (orderId: string) =>
    api
      .post<{ checkout_url: string }>(`/payments/checkout/${orderId}`)
      .then((r) => r.data),
}
