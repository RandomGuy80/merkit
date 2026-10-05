import { forwardRef, InputHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, className, ...props }, ref) => (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">{label}</label>
      )}
      <input
        ref={ref}
        className={cn(
          'w-full rounded-xl border px-4 py-2.5 text-sm text-white placeholder:text-zinc-600 bg-[#111] transition-all duration-150 focus:outline-none focus:ring-1 focus:ring-white/20 focus:border-zinc-600 cursor-none',
          error
            ? 'border-red-900/60 bg-red-950/10'
            : 'border-[#222] hover:border-[#333]',
          className
        )}
        {...props}
      />
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
)
Input.displayName = 'Input'

interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string
  error?: string
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ label, error, className, ...props }, ref) => (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">{label}</label>
      )}
      <textarea
        ref={ref}
        className={cn(
          'w-full rounded-xl border px-4 py-2.5 text-sm text-white placeholder:text-zinc-600 bg-[#111] transition-all duration-150 focus:outline-none focus:ring-1 focus:ring-white/20 focus:border-zinc-600 resize-none cursor-none',
          error
            ? 'border-red-900/60 bg-red-950/10'
            : 'border-[#222] hover:border-[#333]',
          className
        )}
        {...props}
      />
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
)
Textarea.displayName = 'Textarea'
