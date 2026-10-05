export type Role = 'buyer' | 'seller' | 'admin'

export interface User {
  id: string
  email: string
  name: string
  role: Role
  avatar?: string
  bio?: string
  rating: number
  review_count: number
  created_at: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}

export type ListingStatus = 'active' | 'sold' | 'archived'

export interface Listing {
  id: string
  seller_id: string
  category_id?: number
  title: string
  description: string
  price: number
  currency: string
  status: ListingStatus
  location?: string
  images: string[]
  tags: string[]
  views: number
  created_at: string
  updated_at: string
}

export interface ListingsPage {
  items: Listing[]
  next_cursor?: string
}

export interface ListingsFilter {
  q?: string
  category_id?: number
  min_price?: number
  max_price?: number
  seller_id?: string
  cursor?: string
  limit?: number
}

export type OrderStatus =
  | 'pending'
  | 'paid'
  | 'shipped'
  | 'delivered'
  | 'cancelled'
  | 'refunded'

export interface Order {
  id: string
  listing_id: string
  buyer_id: string
  seller_id: string
  amount: number
  platform_fee: number
  status: OrderStatus
  stripe_payment_id?: string
  created_at: string
  updated_at: string
}

export interface Category {
  id: number
  name: string
  slug: string
  icon: string
}

export interface Review {
  id: string
  order_id: string
  reviewer_id: string
  seller_id: string
  rating: number
  body: string
  created_at: string
  reviewer?: { name: string; avatar?: string }
}

export interface ApiError {
  error: string
}
