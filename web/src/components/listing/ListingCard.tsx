'use client'

import Link from 'next/link'
import Image from 'next/image'
import { motion } from 'framer-motion'
import { Eye, MapPin } from 'lucide-react'
import type { Listing } from '@/types'
import { formatPrice, getImageUrl } from '@/lib/utils'
import { ListingStatusBadge } from '@/components/ui/Badge'

export function ListingCard({ listing, index = 0 }: { listing: Listing; index?: number }) {
  const thumb = listing.images?.[0]

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: index * 0.04, ease: [0.25, 0.46, 0.45, 0.94] }}
      whileHover={{ y: -3, transition: { duration: 0.2, ease: 'easeOut' } }}
      className="h-full"
    >
      <Link href={`/listings/${listing.id}`} className="block group h-full">
        <div className="rounded-2xl bg-[#111] border border-[#222] overflow-hidden transition-all duration-200 hover:border-[#333] hover:shadow-xl hover:shadow-black/30 h-full flex flex-col">
          <div className="relative h-48 bg-zinc-900 flex-shrink-0">
            {thumb ? (
              <Image
                src={getImageUrl(thumb)}
                alt={listing.title}
                fill
                className="object-cover group-hover:scale-105 transition-transform duration-500"
                sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                unoptimized={!thumb.startsWith('http')}
              />
            ) : (
              <div className="absolute inset-0 flex items-center justify-center text-4xl text-zinc-700">
                🖼
              </div>
            )}
            <div className="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
            <div className="absolute top-3 right-3">
              <ListingStatusBadge status={listing.status} />
            </div>
          </div>

          <div className="p-4 flex flex-col flex-1">
            <h3 className="font-semibold text-white line-clamp-1 group-hover:text-zinc-200 transition-colors">
              {listing.title}
            </h3>
            <p className="mt-1 text-sm text-zinc-500 line-clamp-2 leading-relaxed flex-1">{listing.description}</p>

            <div className="mt-3 flex items-center justify-between">
              <span className="text-base font-bold text-white">
                {formatPrice(listing.price, listing.currency)}
              </span>
              <div className="flex items-center gap-3 text-xs text-zinc-600">
                {listing.location && (
                  <span className="flex items-center gap-1">
                    <MapPin size={11} /> {listing.location}
                  </span>
                )}
                <span className="flex items-center gap-1">
                  <Eye size={11} /> {listing.views}
                </span>
              </div>
            </div>
          </div>
        </div>
      </Link>
    </motion.div>
  )
}
