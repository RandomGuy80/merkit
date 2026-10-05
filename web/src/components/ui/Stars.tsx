import { Star } from 'lucide-react'
import { cn } from '@/lib/utils'

export function Stars({ rating, max = 5, size = 16 }: { rating: number; max?: number; size?: number }) {
  return (
    <div className="flex items-center gap-0.5">
      {Array.from({ length: max }).map((_, i) => (
        <Star
          key={i}
          size={size}
          className={cn(
            i < Math.round(rating) ? 'fill-white text-white' : 'fill-zinc-700 text-zinc-700'
          )}
        />
      ))}
    </div>
  )
}
