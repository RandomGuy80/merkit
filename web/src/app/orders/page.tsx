'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { motion } from 'framer-motion'
import { ordersApi } from '@/lib/api/orders'
import { OrderStatusBadge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatPrice, formatDate } from '@/lib/utils'
import { useAuthStore } from '@/lib/store/auth'
import { useRouter } from 'next/navigation'
import { useEffect } from 'react'

export default function OrdersPage() {
  const { user, hydrated } = useAuthStore()
  const router = useRouter()

  useEffect(() => {
    if (hydrated && !user) router.push('/login')
  }, [hydrated, user, router])

  const { data: orders, isLoading } = useQuery({
    queryKey: ['orders'],
    queryFn: ordersApi.list,
    enabled: !!user,
  })

  return (
    <div className="max-w-4xl mx-auto px-4 py-10">
      <h1 className="text-xl font-bold text-white tracking-tight mb-7">My Orders</h1>

      {isLoading && (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-20" />)}
        </div>
      )}

      {!isLoading && orders?.length === 0 && (
        <div className="text-center py-24 text-zinc-600">
          <p className="text-5xl mb-4 opacity-30">📦</p>
          <p className="font-medium text-zinc-500">No orders yet</p>
          <Link href="/listings" className="text-zinc-400 text-sm mt-2 inline-block hover:text-white transition-colors">Browse listings →</Link>
        </div>
      )}

      <div className="space-y-2">
        {orders?.map((order, i) => (
          <motion.div
            key={order.id}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.04 }}
          >
            <Link href={`/orders/${order.id}`} className="block bg-[#111] rounded-2xl border border-[#222] p-4 hover:border-[#333] transition-all duration-150">
              <div className="flex items-center justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-zinc-600 mb-1 font-mono">#{order.id.slice(0, 8)}</p>
                  <p className="text-sm font-medium text-zinc-300 truncate">
                    {user?.id === order.buyer_id ? 'Purchased' : 'Sale'}
                  </p>
                </div>
                <div className="text-right flex flex-col items-end gap-1.5">
                  <OrderStatusBadge status={order.status} />
                  <span className="font-semibold text-white text-sm">{formatPrice(order.amount)}</span>
                  <span className="text-xs text-zinc-600">{formatDate(order.created_at)}</span>
                </div>
              </div>
            </Link>
          </motion.div>
        ))}
      </div>
    </div>
  )
}
