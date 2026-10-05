'use client'

import { use } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import Image from 'next/image'
import Link from 'next/link'
import { motion } from 'framer-motion'
import { MapPin, Eye, Calendar, Tag, ShoppingCart, Pencil, Trash2, ArrowLeft, Package } from 'lucide-react'
import { listingsApi } from '@/lib/api/listings'
import { ordersApi } from '@/lib/api/orders'
import { usersApi } from '@/lib/api/users'
import { useAuthStore } from '@/lib/store/auth'
import { Button } from '@/components/ui/Button'
import { Badge, ListingStatusBadge } from '@/components/ui/Badge'
import { Avatar } from '@/components/ui/Avatar'
import { Stars } from '@/components/ui/Stars'
import { Skeleton } from '@/components/ui/Skeleton'
import { useToast } from '@/components/ui/Toast'
import { formatPrice, formatDate, getImageUrl } from '@/lib/utils'
import { useState } from 'react'
import { isAxiosError } from 'axios'
import { DEMO_LISTINGS } from '@/lib/demo'
import { useCartStore } from '@/lib/store/cart'

export default function ListingDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const { user } = useAuthStore()
  const { toast } = useToast()
  const router = useRouter()
  const [buying, setBuying] = useState(false)
  const [imgIdx, setImgIdx] = useState(0)
  const { addItem, hasItem } = useCartStore()

  const isDemo = id.startsWith('demo-')
  const demoListing = isDemo ? DEMO_LISTINGS.find((l) => l.id === id) ?? null : null

  const { data: apiListing, isLoading } = useQuery({
    queryKey: ['listing', id],
    queryFn: () => listingsApi.getById(id),
    enabled: !isDemo,
    retry: false,
  })

  const listing = demoListing ?? apiListing

  const { data: seller } = useQuery({
    queryKey: ['user', listing?.seller_id],
    queryFn: () => usersApi.getPublic(listing!.seller_id),
    enabled: !!listing?.seller_id && !isDemo,
  })

  const handleBuy = async () => {
    if (!user) { router.push('/login'); return }
    setBuying(true)
    try {
      const order = await ordersApi.create(id)
      const { checkout_url } = await ordersApi.checkout(order.id)
      window.location.href = checkout_url
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Failed' : 'Failed'
      toast(msg, 'error')
    } finally {
      setBuying(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm('Delete this listing?')) return
    try {
      await listingsApi.delete(id)
      toast('Listing deleted', 'success')
      router.push('/listings')
    } catch {
      toast('Failed to delete', 'error')
    }
  }

  if (!isDemo && isLoading) {
    return (
      <div className="max-w-5xl mx-auto px-4 py-10">
        <div className="grid md:grid-cols-2 gap-8">
          <Skeleton className="aspect-square rounded-2xl" />
          <div className="space-y-5 pt-2">
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-10 w-1/3" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
            <Skeleton className="h-4 w-2/3" />
            <Skeleton className="h-16 w-full rounded-xl" />
            <Skeleton className="h-12 w-full rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  if (!listing) {
    return (
      <div className="text-center py-24 text-zinc-600 space-y-3">
        <p className="text-5xl opacity-20">📭</p>
        <p className="text-zinc-400 font-medium">Listing not found</p>
        <Link href="/listings" className="inline-flex items-center gap-1.5 text-sm text-zinc-500 hover:text-white transition-colors mt-2">
          <ArrowLeft size={14} /> Back to listings
        </Link>
      </div>
    )
  }

  const images = listing.images?.length ? listing.images : []
  const isOwner = user?.id === listing.seller_id
  const isAdmin = user?.role === 'admin'
  const canBuy = !isDemo && user && !isOwner && listing.status === 'active'

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      {/* Back */}
      <button
        onClick={() => router.back()}
        className="flex items-center gap-1.5 text-xs text-zinc-600 hover:text-zinc-300 transition-colors mb-7 cursor-none"
      >
        <ArrowLeft size={14} /> Back
      </button>

      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.25, 0.46, 0.45, 0.94] }}
        className="grid md:grid-cols-2 gap-10"
      >
        {/* ── Images ── */}
        <div className="space-y-3">
          <div className="relative aspect-square rounded-2xl overflow-hidden bg-zinc-900 border border-[#222]">
            {images[imgIdx] ? (
              <Image
                src={getImageUrl(images[imgIdx])}
                alt={listing.title}
                fill
                className="object-cover"
                sizes="(max-width: 768px) 100vw, 50vw"
                priority
                unoptimized={!images[imgIdx].startsWith('http')}
              />
            ) : (
              <div className="absolute inset-0 flex items-center justify-center text-6xl text-zinc-700">🖼</div>
            )}
          </div>
          {images.length > 1 && (
            <div className="flex gap-2">
              {images.map((img, i) => (
                <button
                  key={i}
                  onClick={() => setImgIdx(i)}
                  className={`relative w-16 h-16 rounded-xl overflow-hidden border-2 transition-all cursor-none ${i === imgIdx ? 'border-white' : 'border-[#222] hover:border-[#444]'}`}
                >
                  <Image src={getImageUrl(img)} alt="" fill className="object-cover" sizes="64px" unoptimized={!img.startsWith('http')} />
                </button>
              ))}
            </div>
          )}
        </div>

        {/* ── Info ── */}
        <div className="flex flex-col gap-5">
          {/* Title + status */}
          <div className="flex items-start justify-between gap-3">
            <h1 className="text-2xl font-bold text-white tracking-tight leading-snug">{listing.title}</h1>
            <ListingStatusBadge status={listing.status} />
          </div>

          {/* Price */}
          <div className="text-4xl font-bold text-white tracking-tight">
            {formatPrice(listing.price, listing.currency)}
          </div>

          {/* Description */}
          <p className="text-zinc-400 leading-relaxed text-sm">{listing.description}</p>

          {/* Meta */}
          <div className="flex flex-wrap gap-4 text-xs text-zinc-600">
            {listing.location && (
              <span className="flex items-center gap-1.5"><MapPin size={12} />{listing.location}</span>
            )}
            <span className="flex items-center gap-1.5"><Eye size={12} />{listing.views} views</span>
            <span className="flex items-center gap-1.5"><Calendar size={12} />{formatDate(listing.created_at)}</span>
          </div>

          {/* Tags */}
          {listing.tags?.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {listing.tags.map((tag) => (
                <Badge key={tag} variant="default">
                  <Tag size={9} className="mr-1" />{tag}
                </Badge>
              ))}
            </div>
          )}

          {/* Seller card */}
          {seller && (
            <Link
              href={`/profile/${seller.id}`}
              className="flex items-center gap-3 p-4 rounded-xl bg-[#111] border border-[#222] hover:border-[#333] transition-colors"
            >
              <Avatar src={seller.avatar} name={seller.name} size={44} />
              <div className="flex-1 min-w-0">
                <p className="font-medium text-white text-sm">{seller.name}</p>
                <div className="flex items-center gap-2 mt-0.5">
                  <Stars rating={seller.rating} size={11} />
                  <span className="text-xs text-zinc-600">{seller.review_count} reviews</span>
                </div>
              </div>
            </Link>
          )}

          {/* Demo seller placeholder */}
          {isDemo && (
            <div className="flex items-center gap-3 p-4 rounded-xl bg-[#111] border border-[#222]">
              <div className="w-11 h-11 rounded-full bg-zinc-800 border border-[#333] flex items-center justify-center text-zinc-500">
                <Package size={18} />
              </div>
              <div>
                <p className="font-medium text-zinc-300 text-sm">Demo Seller</p>
                <div className="flex items-center gap-2 mt-0.5">
                  <Stars rating={4.8} size={11} />
                  <span className="text-xs text-zinc-600">Demo listing</span>
                </div>
              </div>
            </div>
          )}

          {/* ── Actions ── */}
          <div className="flex gap-2 mt-auto pt-1">
            {listing.status === 'active' && !isOwner && (
              <>
                {/* Add to Cart — works for everyone */}
                <Button
                  size="lg"
                  variant={hasItem(listing.id) ? 'secondary' : 'primary'}
                  className="flex-1"
                  onClick={() => {
                    if (!user) { router.push('/login'); return }
                    const added = addItem(listing)
                    if (added) toast('Added to cart', 'success')
                    else toast('Already in cart', 'info')
                  }}
                >
                  <ShoppingCart size={17} />
                  {hasItem(listing.id) ? 'In Cart' : 'Add to Cart'}
                </Button>

                {/* Buy Now — real listings only */}
                {canBuy && (
                  <Button size="lg" variant="secondary" onClick={handleBuy} loading={buying}>
                    Buy Now
                  </Button>
                )}
              </>
            )}

            {/* Owner actions */}
            {isOwner && !isDemo && (
              <Link href={`/listings/${id}/edit`} className="flex-1">
                <Button variant="outline" size="lg" className="w-full">
                  <Pencil size={15} /> Edit
                </Button>
              </Link>
            )}
            {(isOwner || isAdmin) && !isDemo && (
              <Button variant="danger" size="lg" onClick={handleDelete}>
                <Trash2 size={15} />
              </Button>
            )}
          </div>
        </div>
      </motion.div>
    </div>
  )
}
