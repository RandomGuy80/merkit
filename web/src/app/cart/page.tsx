'use client'

import { useState } from 'react'
import Link from 'next/link'
import Image from 'next/image'
import { motion, AnimatePresence } from 'framer-motion'
import { Trash2, ArrowRight, PackageOpen, CreditCard, Lock } from 'lucide-react'
import { useCartStore } from '@/lib/store/cart'
import { useAuthStore } from '@/lib/store/auth'
import { ordersApi } from '@/lib/api/orders'
import { Button } from '@/components/ui/Button'
import { useToast } from '@/components/ui/Toast'
import { formatPrice, getImageUrl } from '@/lib/utils'
import { isAxiosError } from 'axios'
import { useRouter } from 'next/navigation'

export default function CartPage() {
  const { items, removeItem, clearCart } = useCartStore()
  const { user } = useAuthStore()
  const { toast } = useToast()
  const router = useRouter()
  const [checkingOut, setCheckingOut] = useState(false)
  const [buyingOne, setBuyingOne] = useState<string | null>(null)

  const isDemo = (id: string) => id.startsWith('demo-')
  const realItems = items.filter(({ listing }) => !isDemo(listing.id))
  const demoItems = items.filter(({ listing }) => isDemo(listing.id))
  const total = items.reduce((sum, { listing }) => sum + listing.price, 0)
  const realTotal = realItems.reduce((sum, { listing }) => sum + listing.price, 0)

  const handleCheckoutOne = async (listingId: string) => {
    if (!user) { router.push('/login'); return }
    setBuyingOne(listingId)
    try {
      const order = await ordersApi.create(listingId)
      const { checkout_url } = await ordersApi.checkout(order.id)
      window.location.href = checkout_url
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Checkout failed' : 'Checkout failed'
      toast(msg, 'error')
      setBuyingOne(null)
    }
  }

  // Creates orders for all real items, redirects to Stripe for the first one.
  // Remaining orders stay as 'pending' in /orders.
  const handleCheckoutAll = async () => {
    if (!user) { router.push('/login'); return }
    if (realItems.length === 0) {
      toast('Cart only has demo items — create a real listing to test checkout', 'info')
      return
    }
    setCheckingOut(true)
    try {
      const checkoutUrls: string[] = []
      for (const { listing } of realItems) {
        const order = await ordersApi.create(listing.id)
        const { checkout_url } = await ordersApi.checkout(order.id)
        checkoutUrls.push(checkout_url)
      }
      // clear real items from cart; demo items stay
      realItems.forEach(({ listing }) => removeItem(listing.id))
      // go to Stripe for first item; the rest will be pending in /orders
      window.location.href = checkoutUrls[0]
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Checkout failed' : 'Checkout failed'
      toast(msg, 'error')
      setCheckingOut(false)
    }
  }

  if (items.length === 0) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-24 text-center space-y-4">
        <div className="w-16 h-16 rounded-2xl bg-[#111] border border-[#222] flex items-center justify-center mx-auto">
          <PackageOpen size={28} className="text-zinc-600" />
        </div>
        <p className="text-zinc-400 font-medium">Your cart is empty</p>
        <Link href="/listings">
          <Button variant="outline" size="sm" className="mt-2">
            Browse listings <ArrowRight size={14} />
          </Button>
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-10">
      <div className="flex items-center justify-between mb-7">
        <h1 className="text-xl font-bold text-white tracking-tight">
          Cart <span className="text-zinc-600 font-normal text-base ml-1">({items.length})</span>
        </h1>
        <Link href="/listings" className="text-xs text-zinc-500 hover:text-zinc-300 transition-colors flex items-center gap-1">
          Continue shopping <ArrowRight size={12} />
        </Link>
      </div>

      {/* Items */}
      <div className="space-y-2">
        <AnimatePresence initial={false}>
          {items.map(({ listing }) => (
            <motion.div
              key={listing.id}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, x: -24, height: 0, padding: 0, margin: 0, border: 0 }}
              transition={{ duration: 0.2 }}
              className="flex items-center gap-4 bg-[#111] border border-[#222] rounded-2xl p-4"
            >
              {/* Thumbnail */}
              <Link href={`/listings/${listing.id}`} className="relative w-16 h-16 rounded-xl overflow-hidden bg-zinc-900 flex-shrink-0 border border-[#222]">
                {listing.images?.[0] ? (
                  <Image
                    src={getImageUrl(listing.images[0])}
                    alt={listing.title}
                    fill
                    className="object-cover hover:scale-105 transition-transform duration-300"
                    sizes="64px"
                    unoptimized={!listing.images[0].startsWith('http')}
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-zinc-700 text-xl">🖼</div>
                )}
              </Link>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <Link href={`/listings/${listing.id}`} className="text-sm font-medium text-white hover:text-zinc-300 transition-colors line-clamp-1">
                  {listing.title}
                </Link>
                <p className="text-xs text-zinc-600 mt-0.5">
                  {listing.location && `${listing.location} · `}
                  {isDemo(listing.id) ? <span className="text-zinc-700">demo</span> : 'Real listing'}
                </p>
              </div>

              {/* Price + actions */}
              <div className="flex items-center gap-3 flex-shrink-0">
                <span className="font-bold text-white text-sm whitespace-nowrap">
                  {formatPrice(listing.price, listing.currency)}
                </span>

                {!isDemo(listing.id) && (
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => handleCheckoutOne(listing.id)}
                    loading={buyingOne === listing.id}
                  >
                    Pay
                  </Button>
                )}

                <button
                  onClick={() => removeItem(listing.id)}
                  className="text-zinc-700 hover:text-red-400 transition-colors cursor-none"
                >
                  <Trash2 size={15} />
                </button>
              </div>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      {/* Summary + checkout */}
      <div className="mt-5 bg-[#111] border border-[#222] rounded-2xl p-5 space-y-4">
        {/* Totals */}
        <div className="space-y-2">
          {demoItems.length > 0 && realItems.length > 0 && (
            <div className="flex items-center justify-between text-sm">
              <span className="text-zinc-600">Real items ({realItems.length})</span>
              <span className="text-zinc-400">{formatPrice(realTotal)}</span>
            </div>
          )}
          {demoItems.length > 0 && (
            <div className="flex items-center justify-between text-sm">
              <span className="text-zinc-700">Demo items ({demoItems.length})</span>
              <span className="text-zinc-700 line-through">{formatPrice(demoItems.reduce((s, { listing }) => s + listing.price, 0))}</span>
            </div>
          )}
          <div className="flex items-center justify-between pt-2 border-t border-[#222]">
            <span className="text-zinc-300 font-medium">Total</span>
            <span className="font-bold text-white text-xl">{formatPrice(realTotal || total)}</span>
          </div>
        </div>

        {/* Pay all button — always visible */}
        <Button
          size="lg"
          className="w-full"
          loading={checkingOut}
          onClick={handleCheckoutAll}
        >
          <CreditCard size={17} />
          {checkingOut
            ? 'Creating orders…'
            : realItems.length > 0
              ? `Pay All — ${formatPrice(realTotal)}`
              : 'Pay All'}
        </Button>

        {realItems.length > 1 && (
          <p className="text-xs text-zinc-700 text-center">
            Each item is a separate Stripe session — you&apos;ll pay the first item now, remaining orders stay pending in{' '}
            <Link href="/orders" className="text-zinc-500 hover:text-zinc-300 transition-colors">Orders</Link>.
          </p>
        )}

        <div className="flex items-center justify-center gap-1.5 text-xs text-zinc-700">
          <Lock size={11} />
          Secured by Stripe
        </div>
      </div>
    </div>
  )
}
