'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ShoppingBag, ShoppingCart, Plus, User, LogOut, Shield, Package } from 'lucide-react'
import { useAuthStore } from '@/lib/store/auth'
import { useAuth } from '@/lib/hooks/useAuth'
import { useCartStore } from '@/lib/store/cart'
import { Avatar } from '@/components/ui/Avatar'
import { Button } from '@/components/ui/Button'
import { cn } from '@/lib/utils'
import { useState } from 'react'

export function Header() {
  const { user } = useAuthStore()
  const { logout } = useAuth()
  const { items } = useCartStore()
  const pathname = usePathname()
  const [menuOpen, setMenuOpen] = useState(false)

  const cartCount = items.length

  const nav = [
    { href: '/listings', label: 'Browse' },
    ...(user ? [{ href: '/orders', label: 'Orders' }] : []),
    ...(user?.role === 'admin' ? [{ href: '/admin', label: 'Admin' }] : []),
  ]

  return (
    <header className="sticky top-0 z-40 bg-[#0a0a0a]/80 backdrop-blur-xl border-b border-[#222]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
        <Link href="/" className="flex items-center gap-2 font-bold text-lg text-white tracking-tight">
          <ShoppingBag size={20} className="text-zinc-400" />
          <span>Merkit</span>
        </Link>

        <nav className="hidden md:flex items-center gap-1">
          {nav.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'px-3 py-1.5 rounded-lg text-sm font-medium transition-all duration-150',
                pathname.startsWith(item.href)
                  ? 'bg-white/8 text-white'
                  : 'text-zinc-400 hover:text-white hover:bg-white/5'
              )}
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          {/* Cart icon — visible to everyone */}
          <Link
            href="/cart"
            className="relative p-2 rounded-xl text-zinc-400 hover:text-white hover:bg-white/5 transition-colors cursor-none"
          >
            <ShoppingCart size={20} />
            {cartCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-white text-black text-[10px] font-bold flex items-center justify-center leading-none">
                {cartCount > 9 ? '9+' : cartCount}
              </span>
            )}
          </Link>

          {user ? (
            <>
              {(user.role === 'seller' || user.role === 'admin') && (
                <Link href="/listings/new">
                  <Button size="sm" variant="secondary">
                    <Plus size={14} /> New Listing
                  </Button>
                </Link>
              )}
              <div className="relative">
                <button
                  onClick={() => setMenuOpen(!menuOpen)}
                  className="flex items-center gap-2 px-2 py-1.5 rounded-xl hover:bg-white/5 transition-colors cursor-none"
                >
                  <Avatar src={user.avatar} name={user.name} size={30} />
                  <span className="hidden sm:block text-sm font-medium text-zinc-300">{user.name}</span>
                </button>
                {menuOpen && (
                  <div
                    className="absolute right-0 top-full mt-2 w-48 bg-[#111] rounded-2xl border border-[#2e2e2e] shadow-2xl shadow-black/60 py-1 z-50"
                    onMouseLeave={() => setMenuOpen(false)}
                  >
                    <Link
                      href="/profile/me"
                      className="flex items-center gap-2.5 px-4 py-2.5 text-sm text-zinc-300 hover:text-white hover:bg-white/5 transition-colors"
                      onClick={() => setMenuOpen(false)}
                    >
                      <User size={14} className="text-zinc-500" /> Profile
                    </Link>
                    <Link
                      href="/orders"
                      className="flex items-center gap-2.5 px-4 py-2.5 text-sm text-zinc-300 hover:text-white hover:bg-white/5 transition-colors"
                      onClick={() => setMenuOpen(false)}
                    >
                      <Package size={14} className="text-zinc-500" /> Orders
                    </Link>
                    {user.role === 'admin' && (
                      <Link
                        href="/admin"
                        className="flex items-center gap-2.5 px-4 py-2.5 text-sm text-zinc-300 hover:text-white hover:bg-white/5 transition-colors"
                        onClick={() => setMenuOpen(false)}
                      >
                        <Shield size={14} className="text-zinc-500" /> Admin
                      </Link>
                    )}
                    <div className="my-1 border-t border-[#222]" />
                    <button
                      onClick={() => { setMenuOpen(false); logout() }}
                      className="flex items-center gap-2.5 w-full px-4 py-2.5 text-sm text-red-400 hover:bg-red-950/30 transition-colors cursor-none"
                    >
                      <LogOut size={14} /> Sign out
                    </button>
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="flex items-center gap-2">
              <Link href="/login">
                <Button variant="ghost" size="sm">Sign in</Button>
              </Link>
              <Link href="/register">
                <Button size="sm">Sign up</Button>
              </Link>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
