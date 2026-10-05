'use client'

import { use } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { CreditCard, CheckCircle, Truck, Package, XCircle } from 'lucide-react'
import { ordersApi } from '@/lib/api/orders'
import { OrderStatusBadge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { useToast } from '@/components/ui/Toast'
import { useAuthStore } from '@/lib/store/auth'
import { formatPrice, formatDate } from '@/lib/utils'
import type { OrderStatus } from '@/types'
import { useState } from 'react'
import { isAxiosError } from 'axios'

const STATUS_STEPS: OrderStatus[] = ['pending', 'paid', 'shipped', 'delivered']

export default function OrderDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const { user } = useAuthStore()
  const { toast } = useToast()
  const qc = useQueryClient()
  const [paying, setPaying] = useState(false)

  const { data: order, isLoading } = useQuery({
    queryKey: ['order', id],
    queryFn: () => ordersApi.getById(id),
  })

  const updateStatus = async (status: OrderStatus) => {
    try {
      await ordersApi.updateStatus(id, status)
      qc.invalidateQueries({ queryKey: ['order', id] })
      qc.invalidateQueries({ queryKey: ['orders'] })
      toast(`Status updated to ${status}`, 'success')
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Failed' : 'Failed'
      toast(msg, 'error')
    }
  }

  const handlePay = async () => {
    setPaying(true)
    try {
      const { checkout_url } = await ordersApi.checkout(id)
      window.location.href = checkout_url
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Failed' : 'Failed'
      toast(msg, 'error')
      setPaying(false)
    }
  }

  if (isLoading) return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-4">
      <Skeleton className="h-48" /><Skeleton className="h-24" />
    </div>
  )
  if (!order) return <div className="text-center py-24 text-zinc-600">Order not found</div>

  const isBuyer = user?.id === order.buyer_id
  const isSeller = user?.id === order.seller_id
  const isAdmin = user?.role === 'admin'

  const stepIdx = STATUS_STEPS.indexOf(order.status as OrderStatus)

  return (
    <div className="max-w-2xl mx-auto px-4 py-10">
      <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.35 }} className="space-y-4">

        {/* Header */}
        <div className="bg-[#111] rounded-2xl border border-[#222] p-6">
          <div className="flex items-center justify-between mb-5">
            <div>
              <p className="text-xs text-zinc-600 uppercase tracking-wider mb-1">Order ID</p>
              <p className="font-mono text-sm text-zinc-400">{order.id}</p>
            </div>
            <OrderStatusBadge status={order.status} />
          </div>
          <div className="grid grid-cols-2 gap-5 text-sm">
            <div>
              <p className="text-zinc-600 text-xs uppercase tracking-wider mb-1">Amount</p>
              <p className="font-bold text-2xl text-white">{formatPrice(order.amount)}</p>
            </div>
            <div>
              <p className="text-zinc-600 text-xs uppercase tracking-wider mb-1">Platform fee</p>
              <p className="font-medium text-zinc-300">{formatPrice(order.platform_fee)}</p>
            </div>
            <div>
              <p className="text-zinc-600 text-xs uppercase tracking-wider mb-1">Created</p>
              <p className="font-medium text-zinc-300">{formatDate(order.created_at)}</p>
            </div>
            <div>
              <p className="text-zinc-600 text-xs uppercase tracking-wider mb-1">Updated</p>
              <p className="font-medium text-zinc-300">{formatDate(order.updated_at)}</p>
            </div>
          </div>
        </div>

        {/* Progress */}
        {order.status !== 'cancelled' && order.status !== 'refunded' && (
          <div className="bg-[#111] rounded-2xl border border-[#222] p-6">
            <h2 className="text-xs font-medium text-zinc-500 uppercase tracking-wider mb-5">Progress</h2>
            <div className="flex items-center">
              {STATUS_STEPS.map((step, i) => (
                <div key={step} className="flex items-center flex-1">
                  <div className="flex flex-col items-center gap-1.5">
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold border transition-colors ${i <= stepIdx ? 'bg-white border-white text-black' : 'border-[#333] text-zinc-600'}`}>
                      {i < stepIdx ? <CheckCircle size={14} /> : i + 1}
                    </div>
                    <span className={`text-xs capitalize whitespace-nowrap transition-colors ${i <= stepIdx ? 'text-zinc-300' : 'text-zinc-700'}`}>{step}</span>
                  </div>
                  {i < STATUS_STEPS.length - 1 && (
                    <div className={`flex-1 h-px mx-2 transition-colors ${i < stepIdx ? 'bg-white/30' : 'bg-[#2e2e2e]'}`} />
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="bg-[#111] rounded-2xl border border-[#222] p-6 space-y-3">
          <h2 className="text-xs font-medium text-zinc-500 uppercase tracking-wider mb-4">Actions</h2>

          {isBuyer && order.status === 'pending' && (
            <Button className="w-full" size="lg" onClick={handlePay} loading={paying}>
              <CreditCard size={17} /> Pay Now
            </Button>
          )}
          {isSeller && order.status === 'paid' && (
            <Button className="w-full" variant="secondary" onClick={() => updateStatus('shipped')}>
              <Truck size={15} /> Mark as Shipped
            </Button>
          )}
          {(isBuyer || isSeller) && order.status === 'shipped' && (
            <Button className="w-full" variant="secondary" onClick={() => updateStatus('delivered')}>
              <Package size={15} /> Mark as Delivered
            </Button>
          )}
          {(isBuyer || isSeller) && order.status === 'pending' && (
            <Button variant="outline" className="w-full" onClick={() => updateStatus('cancelled')}>
              <XCircle size={15} /> Cancel Order
            </Button>
          )}
          {isAdmin && order.status === 'delivered' && (
            <Button variant="danger" className="w-full" onClick={() => updateStatus('refunded')}>
              Issue Refund
            </Button>
          )}
          {!isBuyer && !isSeller && !isAdmin && (
            <p className="text-sm text-zinc-600 text-center">No actions available</p>
          )}
        </div>
      </motion.div>
    </div>
  )
}
