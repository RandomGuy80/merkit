import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Listing } from '@/types'

interface CartItem {
  listing: Listing
}

interface CartState {
  items: CartItem[]
  addItem: (listing: Listing) => boolean  // returns false if already in cart
  removeItem: (id: string) => void
  clearCart: () => void
  hasItem: (id: string) => boolean
}

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],
      addItem: (listing) => {
        if (get().items.some((i) => i.listing.id === listing.id)) return false
        set((s) => ({ items: [...s.items, { listing }] }))
        return true
      },
      removeItem: (id) =>
        set((s) => ({ items: s.items.filter((i) => i.listing.id !== id) })),
      clearCart: () => set({ items: [] }),
      hasItem: (id) => get().items.some((i) => i.listing.id === id),
    }),
    { name: 'merkit-cart' }
  )
)
