'use client'

import { use } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { usersApi } from '@/lib/api/users'
import { listingsApi } from '@/lib/api/listings'
import { Avatar } from '@/components/ui/Avatar'
import { Stars } from '@/components/ui/Stars'
import { Badge } from '@/components/ui/Badge'
import { ListingCard } from '@/components/listing/ListingCard'
import { Skeleton } from '@/components/ui/Skeleton'
import { formatDate } from '@/lib/utils'

export default function PublicProfilePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)

  const { data: user, isLoading } = useQuery({
    queryKey: ['user', id],
    queryFn: () => usersApi.getPublic(id),
  })

  const { data: listings } = useQuery({
    queryKey: ['listings', 'seller', id],
    queryFn: () => listingsApi.search({ seller_id: id, limit: 12 }),
    enabled: !!user,
  })

  const { data: reviews } = useQuery({
    queryKey: ['reviews', id],
    queryFn: () => usersApi.getReviews(id),
    enabled: !!user,
  })

  if (isLoading) return (
    <div className="max-w-4xl mx-auto px-4 py-10 space-y-4">
      <Skeleton className="h-32" /><Skeleton className="h-48" />
    </div>
  )
  if (!user) return <div className="text-center py-24 text-zinc-600">User not found</div>

  return (
    <div className="max-w-4xl mx-auto px-4 py-10 space-y-8">
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35 }}
        className="bg-[#111] rounded-2xl border border-[#222] p-6 flex items-start gap-6"
      >
        <Avatar src={user.avatar} name={user.name} size={80} />
        <div className="flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-2xl font-bold text-white tracking-tight">{user.name}</h1>
            <Badge variant={user.role === 'seller' ? 'purple' : 'info'}>{user.role}</Badge>
          </div>
          {user.bio && <p className="text-zinc-400 mt-2 text-sm leading-relaxed">{user.bio}</p>}
          <div className="flex items-center gap-5 mt-3">
            {user.role === 'seller' && (
              <>
                <div className="flex items-center gap-2">
                  <Stars rating={user.rating} size={14} />
                  <span className="text-sm font-medium text-zinc-300">{user.rating.toFixed(1)}</span>
                </div>
                <span className="text-sm text-zinc-600">{user.review_count} reviews</span>
              </>
            )}
            <span className="text-sm text-zinc-700">Since {formatDate(user.created_at)}</span>
          </div>
        </div>
      </motion.div>

      {listings && listings.items.length > 0 && (
        <section>
          <h2 className="text-xs font-medium text-zinc-500 uppercase tracking-wider mb-5">Listings</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {listings.items.map((l, i) => <ListingCard key={l.id} listing={l} index={i} />)}
          </div>
        </section>
      )}

      {reviews && reviews.length > 0 && (
        <section>
          <h2 className="text-xs font-medium text-zinc-500 uppercase tracking-wider mb-5">Reviews</h2>
          <div className="space-y-3">
            {reviews.map((r) => (
              <motion.div
                key={r.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                className="bg-[#111] rounded-2xl border border-[#222] p-4"
              >
                <div className="flex items-center gap-3 mb-3">
                  <Avatar src={r.reviewer?.avatar} name={r.reviewer?.name ?? '?'} size={34} />
                  <div>
                    <p className="text-sm font-medium text-zinc-200">{r.reviewer?.name}</p>
                    <Stars rating={r.rating} size={12} />
                  </div>
                  <span className="ml-auto text-xs text-zinc-700">{formatDate(r.created_at)}</span>
                </div>
                {r.body && <p className="text-sm text-zinc-400 leading-relaxed">{r.body}</p>}
              </motion.div>
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
