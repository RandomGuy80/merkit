'use client'

import { Suspense, useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search } from 'lucide-react'
import { listingsApi } from '@/lib/api/listings'
import { categoriesApi } from '@/lib/api/users'
import { ListingCard } from '@/components/listing/ListingCard'
import { ListingCardSkeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/Button'
import type { Listing, ListingsFilter } from '@/types'
import { useSearchParams, useRouter } from 'next/navigation'
import { DEMO_LISTINGS } from '@/lib/demo'

function ListingsContent() {
  const searchParams = useSearchParams()
  const router = useRouter()

  const [filter, setFilter] = useState<ListingsFilter>({
    q: searchParams.get('q') ?? '',
    category_id: searchParams.get('category_id') ? Number(searchParams.get('category_id')) : undefined,
    limit: 20,
  })
  const [input, setInput] = useState(filter.q ?? '')
  const [cursor, setCursor] = useState<string | undefined>()
  const [allItems, setAllItems] = useState<Listing[]>([])

  // sync filter when URL changes (e.g. navigating from home page categories)
  useEffect(() => {
    const q = searchParams.get('q') ?? ''
    const catId = searchParams.get('category_id') ? Number(searchParams.get('category_id')) : undefined
    setInput(q)
    setCursor(undefined)
    setAllItems([])
    setFilter({ q, category_id: catId, limit: 20 })
  }, [searchParams])

  const { data, isLoading } = useQuery({
    queryKey: ['listings', filter, cursor],
    queryFn: () => listingsApi.search({ ...filter, cursor }),
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: categoriesApi.list,
  })

  useEffect(() => {
    if (data) {
      const items = data.items ?? []
      if (!cursor) setAllItems(items)
      else setAllItems((prev) => [...prev, ...items])
    }
  }, [data, cursor])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setCursor(undefined)
    setAllItems([])
    setFilter((f) => ({ ...f, q: input }))
    router.push(`/listings?q=${encodeURIComponent(input)}`)
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <form onSubmit={handleSearch} className="flex gap-2 mb-8">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-600" size={17} />
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Search listings..."
            className="w-full pl-11 pr-4 py-2.5 rounded-xl border border-[#222] text-sm bg-[#111] text-white placeholder:text-zinc-600 focus:outline-none focus:border-[#333] focus:ring-1 focus:ring-white/10 transition-all cursor-none"
          />
        </div>
        <Button type="submit" size="md">Search</Button>
      </form>

      {categories && (
        <div className="flex gap-2 flex-wrap mb-8">
          <button
            onClick={() => { setCursor(undefined); setAllItems([]); setFilter((f) => ({ ...f, category_id: undefined })) }}
            className={`px-3 py-1.5 rounded-full text-xs font-medium transition-all duration-150 cursor-none ${!filter.category_id ? 'bg-white text-black' : 'bg-[#111] border border-[#222] text-zinc-500 hover:border-[#333] hover:text-zinc-300'}`}
          >
            All
          </button>
          {categories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => { setCursor(undefined); setAllItems([]); setFilter((f) => ({ ...f, category_id: cat.id })) }}
              className={`px-3 py-1.5 rounded-full text-xs font-medium transition-all duration-150 cursor-none ${filter.category_id === cat.id ? 'bg-white text-black' : 'bg-[#111] border border-[#222] text-zinc-500 hover:border-[#333] hover:text-zinc-300'}`}
            >
              {cat.icon} {cat.name}
            </button>
          ))}
        </div>
      )}

      {(() => {
        const filteredDemo = DEMO_LISTINGS.filter((l) => {
          const matchCat = !filter.category_id || l.category_id === filter.category_id
          const q = filter.q?.toLowerCase() ?? ''
          const matchQ = !q || l.title.toLowerCase().includes(q) || l.description.toLowerCase().includes(q)
          return matchCat && matchQ
        })
        const items = [...allItems, ...filteredDemo]

        return (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {isLoading && !allItems.length
                ? Array.from({ length: 8 }).map((_, i) => <ListingCardSkeleton key={i} />)
                : items.map((l, i) => <ListingCard key={l.id} listing={l} index={i} />)}
            </div>

            {data?.next_cursor && (
              <div className="mt-10 text-center">
                <Button variant="outline" onClick={() => setCursor(data.next_cursor)}>
                  Load more
                </Button>
              </div>
            )}

            {!isLoading && items.length === 0 && (
              <div className="text-center py-24 text-zinc-600">
                <p className="text-5xl mb-4 opacity-30">🔍</p>
                <p className="text-base font-medium text-zinc-500">No listings found</p>
                <p className="text-sm text-zinc-700 mt-1">Try a different search or category</p>
              </div>
            )}
          </>
        )
      })()}
    </div>
  )
}

export default function ListingsPage() {
  return (
    <Suspense fallback={
      <div className="max-w-7xl mx-auto px-4 py-10 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {Array.from({ length: 8 }).map((_, i) => <ListingCardSkeleton key={i} />)}
      </div>
    }>
      <ListingsContent />
    </Suspense>
  )
}
