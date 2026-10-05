import Image from 'next/image'
import { cn } from '@/lib/utils'

interface AvatarProps {
  src?: string | null
  name: string
  size?: number
  className?: string
}

export function Avatar({ src, name, size = 40, className }: AvatarProps) {
  const initials = name
    .split(' ')
    .map((w) => w[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()

  if (src) {
    const isExternal = src.startsWith('http')
    const fullSrc = isExternal ? src : `http://localhost:8080${src}`
    return (
      <div
        className={cn('relative rounded-full overflow-hidden flex-shrink-0', className)}
        style={{ width: size, height: size }}
      >
        {/* unoptimized for backend images — bypasses Next.js image CDN cache */}
        <Image src={fullSrc} alt={name} fill className="object-cover" sizes={`${size}px`} unoptimized={!isExternal} />
      </div>
    )
  }

  return (
    <div
      className={cn(
        'rounded-full bg-zinc-800 text-zinc-200 font-semibold flex items-center justify-center flex-shrink-0 border border-[#333]',
        className
      )}
      style={{ width: size, height: size, fontSize: size * 0.38 }}
    >
      {initials}
    </div>
  )
}
