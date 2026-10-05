'use client'

import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import api from '@/lib/api/client'
import { useAuthStore } from '@/lib/store/auth'
import { Avatar } from '@/components/ui/Avatar'
import { Badge, OrderStatusBadge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatPrice, formatDate } from '@/lib/utils'
import type { User, Order } from '@/types'

export default function AdminPage() {
  const { user } = useAuthStore()
  const router = useRouter()

  useEffect(() => {
    if (user && user.role !== 'admin') router.push('/')
    if (user === null) router.push('/login')
  }, [user, router])

  const { data: users, isLoading: loadingUsers } = useQuery({
    queryKey: ['admin', 'users'],
    queryFn: () => api.get<User[]>('/admin/users?limit=50').then((r) => r.data),
    enabled: user?.role === 'admin',
  })

  const { data: orders, isLoading: loadingOrders } = useQuery({
    queryKey: ['admin', 'orders'],
    queryFn: () => api.get<Order[]>('/admin/orders?limit=50').then((r) => r.data),
    enabled: user?.role === 'admin',
  })

  if (!user || user.role !== 'admin') return null

  const stats = [
    { label: 'Users', value: users?.length ?? '—' },
    { label: 'Orders', value: orders?.length ?? '—' },
    { label: 'Revenue', value: orders ? formatPrice(orders.reduce((s, o) => s + o.platform_fee, 0)) : '—' },
    { label: 'Active', value: orders?.filter((o) => o.status !== 'cancelled' && o.status !== 'refunded').length ?? '—' },
  ]

  return (
    <div className="max-w-6xl mx-auto px-4 py-10 space-y-10">
      <h1 className="text-xl font-bold text-white tracking-tight">Admin Dashboard</h1>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {stats.map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.06 }}
            className="bg-[#111] rounded-2xl border border-[#222] p-5 text-center"
          >
            <p className="text-2xl font-bold text-white">{stat.value}</p>
            <p className="text-xs text-zinc-600 mt-1 uppercase tracking-wider">{stat.label}</p>
          </motion.div>
        ))}
      </div>

      <section>
        <h2 className="text-xs font-medium text-zinc-500 uppercase tracking-wider mb-4">Users</h2>
        {loadingUsers ? <Skeleton className="h-40" /> : (
          <div className="bg-[#111] rounded-2xl border border-[#222] overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b border-[#222]">
                  <tr>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">User</th>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">Role</th>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">Rating</th>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">Joined</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#1a1a1a]">
                  {users?.map((u) => (
                    <motion.tr key={u.id} initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="hover:bg-[#1a1a1a] transition-colors">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          <Avatar src={u.avatar} name={u.name} size={30} />
                          <div>
                            <p className="font-medium text-zinc-200 text-sm">{u.name}</p>
                            <p className="text-xs text-zinc-600">{u.email}</p>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <Badge variant={u.role === 'admin' ? 'danger' : u.role === 'seller' ? 'purple' : 'info'}>{u.role}</Badge>
                      </td>
                      <td className="px-4 py-3 text-zinc-400 text-sm">{u.rating.toFixed(1)} ({u.review_count})</td>
                      <td className="px-4 py-3 text-zinc-600 text-sm">{formatDate(u.created_at)}</td>
                    </motion.tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </section>

      <section>
        <h2 className="text-xs font-medium text-zinc-500 uppercase tracking-wider mb-4">Orders</h2>
        {loadingOrders ? <Skeleton className="h-40" /> : (
          <div className="bg-[#111] rounded-2xl border border-[#222] overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b border-[#222]">
                  <tr>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">ID</th>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">Status</th>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">Amount</th>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">Fee</th>
                    <th className="text-left px-4 py-3 text-xs text-zinc-600 font-medium uppercase tracking-wider">Date</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#1a1a1a]">
                  {orders?.map((o) => (
                    <tr key={o.id} className="hover:bg-[#1a1a1a] transition-colors">
                      <td className="px-4 py-3 font-mono text-xs text-zinc-600">{o.id.slice(0, 8)}</td>
                      <td className="px-4 py-3"><OrderStatusBadge status={o.status} /></td>
                      <td className="px-4 py-3 font-medium text-zinc-200">{formatPrice(o.amount)}</td>
                      <td className="px-4 py-3 text-zinc-500">{formatPrice(o.platform_fee)}</td>
                      <td className="px-4 py-3 text-zinc-600">{formatDate(o.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </section>
    </div>
  )
}
