import Link from 'next/link'
import { ShoppingBag } from 'lucide-react'

export function Footer() {
  return (
    <footer className="border-t border-[#222] mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 flex flex-col sm:flex-row items-center justify-between gap-4">
        <Link href="/" className="flex items-center gap-2 font-bold text-base text-white tracking-tight">
          <ShoppingBag size={18} className="text-zinc-500" />
          <span>Merkit</span>
        </Link>
        <p className="text-sm text-zinc-600">© 2026 Merkit. Portfolio project.</p>
        <div className="flex items-center gap-5 text-sm text-zinc-600">
          <Link href="/listings" className="hover:text-zinc-300 transition-colors">Browse</Link>
          <Link href="/register" className="hover:text-zinc-300 transition-colors">Sell</Link>
        </div>
      </div>
    </footer>
  )
}
