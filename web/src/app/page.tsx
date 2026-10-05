'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Search, ArrowRight, ShieldCheck, Zap, Star } from 'lucide-react'
import { listingsApi } from '@/lib/api/listings'
import { categoriesApi } from '@/lib/api/users'
import { ListingCard } from '@/components/listing/ListingCard'
import { ListingCardSkeleton } from '@/components/ui/Skeleton'
import { Button } from '@/components/ui/Button'
import { DEMO_LISTINGS } from '@/lib/demo'

export default function HomePage() {
  const [query, setQuery] = useState('')
  const router = useRouter()

  const { data: featured, isLoading } = useQuery({
    queryKey: ['listings', 'featured'],
    queryFn: () => listingsApi.search({ limit: 8 }),
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: categoriesApi.list,
  })

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    router.push(`/listings?q=${encodeURIComponent(query)}`)
  }

  return (
    <div className="space-y-20">
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_50%_0%,rgba(255,255,255,0.04),transparent_70%)]" />
        <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />
        <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-28 text-center">
          <motion.div
            initial={{ opacity: 0, y: 32 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] }}
          >
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-[#2e2e2e] bg-white/3 text-xs text-zinc-400 mb-6 tracking-wide">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              Freelance marketplace — open for business
            </div>
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold leading-tight tracking-tight text-white">
              Find the perfect<br />
              <span className="text-zinc-400">freelance service</span>
            </h1>
            <p className="mt-5 text-base sm:text-lg text-zinc-500 max-w-xl mx-auto leading-relaxed">
              Buy and sell digital services. Connect with talented professionals worldwide.
            </p>

            <form onSubmit={handleSearch} className="mt-8 flex max-w-lg mx-auto gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-zinc-600" size={18} />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search for any service..."
                  className="w-full pl-11 pr-4 py-3 rounded-xl text-white bg-[#111] border border-[#222] text-sm focus:outline-none focus:border-[#333] focus:ring-1 focus:ring-white/10 transition-all placeholder:text-zinc-600 cursor-none"
                />
              </div>
              <Button type="submit" size="lg">Search</Button>
            </form>

            <p className="mt-4 text-xs text-zinc-600">Popular: logo design · web development · video editing</p>
          </motion.div>
        </div>
      </section>

      {/* Categories */}
      {categories && (
        <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-sm font-medium text-zinc-500 uppercase tracking-wider mb-5">Browse Categories</h2>
          <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3">
            {categories.map((cat, i) => (
              <motion.div
                key={cat.id}
                initial={{ opacity: 0, scale: 0.92 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: i * 0.04, duration: 0.3 }}
              >
                <Link
                  href={`/listings?category_id=${cat.id}`}
                  className="flex flex-col items-center gap-2 p-4 bg-[#111] rounded-2xl border border-[#222] hover:border-[#333] hover:bg-[#1a1a1a] hover:-translate-y-1 transition-all duration-200 text-center group"
                >
                  <span className="text-2xl">{cat.icon}</span>
                  <span className="text-xs font-medium text-zinc-500 group-hover:text-zinc-300 leading-tight transition-colors">{cat.name}</span>
                </Link>
              </motion.div>
            ))}
          </div>
        </section>
      )}

      {/* Featured listings */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-sm font-medium text-zinc-500 uppercase tracking-wider">Featured Listings</h2>
          <Link href="/listings" className="flex items-center gap-1 text-xs text-zinc-400 hover:text-white transition-colors font-medium">
            See all <ArrowRight size={13} />
          </Link>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {isLoading
            ? Array.from({ length: 8 }).map((_, i) => <ListingCardSkeleton key={i} />)
            : [...(featured?.items ?? []), ...DEMO_LISTINGS].map((l, i) => (
                <ListingCard key={l.id} listing={l} index={i} />
              ))}
        </div>
      </section>

      {/* Features */}
      <section className="border-y border-[#222] bg-[#0d0d0d]">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-20 grid sm:grid-cols-3 gap-10">
          {[
            { icon: <ShieldCheck size={24} />, title: 'Secure Payments', desc: 'Stripe-powered checkout with buyer protection and refund policies.' },
            { icon: <Zap size={24} />, title: 'Real-time Updates', desc: 'WebSocket notifications on every order event, instantly.' },
            { icon: <Star size={24} />, title: 'Verified Reviews', desc: 'Only buyers of delivered orders can leave reviews.' },
          ].map((item, i) => (
            <motion.div
              key={item.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1, duration: 0.4 }}
              className="flex flex-col gap-3"
            >
              <div className="w-10 h-10 rounded-xl bg-white/5 border border-[#2e2e2e] flex items-center justify-center text-zinc-400">
                {item.icon}
              </div>
              <h3 className="font-semibold text-white text-sm">{item.title}</h3>
              <p className="text-sm text-zinc-500 leading-relaxed">{item.desc}</p>
            </motion.div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="max-w-2xl mx-auto px-4 pb-20 text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
        >
          <h2 className="text-2xl font-bold text-white tracking-tight mb-3">Ready to start selling?</h2>
          <p className="text-zinc-500 mb-7 text-sm">Create your seller account and publish your first listing in minutes.</p>
          <Link href="/register">
            <Button size="lg">Get started — it&apos;s free</Button>
          </Link>
        </motion.div>
      </section>
    </div>
  )
}
