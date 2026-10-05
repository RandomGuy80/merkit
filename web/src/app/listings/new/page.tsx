'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useRouter } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { useState, useRef } from 'react'
import { Upload, X } from 'lucide-react'
import { Input, Textarea } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { useToast } from '@/components/ui/Toast'
import { listingsApi } from '@/lib/api/listings'
import { categoriesApi } from '@/lib/api/users'
import { useAuthStore } from '@/lib/store/auth'
import { isAxiosError } from 'axios'

const schema = z.object({
  title: z.string().min(3, 'Min 3 characters'),
  description: z.string().min(10, 'Min 10 characters'),
  price: z.string().min(1, 'Required'),
  category_id: z.string().optional(),
  location: z.string().optional(),
  tags: z.string().optional(),
})
type FormData = z.infer<typeof schema>

export default function NewListingPage() {
  const { user } = useAuthStore()
  const { toast } = useToast()
  const router = useRouter()
  const [images, setImages] = useState<File[]>([])
  const [previews, setPreviews] = useState<string[]>([])
  const fileRef = useRef<HTMLInputElement>(null)

  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: categoriesApi.list })

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm({
    resolver: zodResolver(schema),
  })

  if (!user || (user.role !== 'seller' && user.role !== 'admin')) {
    return (
      <div className="max-w-lg mx-auto px-4 py-24 text-center text-zinc-600">
        <p className="text-4xl mb-4 opacity-30">🏪</p>
        <p className="font-medium text-zinc-400">Only sellers can create listings.</p>
      </div>
    )
  }

  const addImage = (files: FileList | null) => {
    if (!files) return
    const newFiles = Array.from(files).slice(0, 5 - images.length)
    const tooBig = newFiles.filter((f) => f.size > 10 * 1024 * 1024)
    if (tooBig.length) { toast('Images must be under 10MB', 'error'); return }
    setImages((prev) => [...prev, ...newFiles])
    newFiles.forEach((f) => {
      const reader = new FileReader()
      reader.onload = (e) => setPreviews((prev) => [...prev, e.target?.result as string])
      reader.readAsDataURL(f)
    })
  }

  const removeImage = (i: number) => {
    setImages((prev) => prev.filter((_, idx) => idx !== i))
    setPreviews((prev) => prev.filter((_, idx) => idx !== i))
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onSubmit = async (data: any) => {
    const price = parseFloat(data.price)
    if (isNaN(price) || price <= 0) { toast('Price must be a positive number', 'error'); return }
    try {
      const listing = await listingsApi.create({
        title: data.title,
        description: data.description,
        price,
        currency: data.currency || 'USD',
        category_id: data.category_id ? parseInt(data.category_id, 10) : undefined,
        location: data.location || undefined,
        tags: data.tags ? data.tags.split(',').map((t: string) => t.trim()).filter(Boolean) : [],
      })
      for (const img of images) {
        await listingsApi.addImage(listing.id, img)
      }
      toast('Listing created!', 'success')
      router.push(`/listings/${listing.id}`)
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Failed' : 'Failed'
      toast(msg, 'error')
    }
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-10">
      <h1 className="text-xl font-bold text-white tracking-tight mb-7">New Listing</h1>
      <form onSubmit={handleSubmit(onSubmit)} className="bg-[#111] rounded-2xl border border-[#222] p-6 space-y-5">
        <Input label="Title" placeholder="e.g. Logo Design Service" error={errors.title?.message as string | undefined} {...register('title')} />
        <Textarea label="Description" rows={4} placeholder="Describe your service in detail..." error={errors.description?.message as string | undefined} {...register('description')} />

        <div className="grid grid-cols-2 gap-4">
          <Input label="Price (USD)" type="number" step="0.01" placeholder="0.00" error={errors.price?.message as string | undefined} {...register('price')} />
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Category</label>
            <select
              {...register('category_id')}
              className="w-full rounded-xl border border-[#222] px-4 py-2.5 text-sm bg-[#111] text-white focus:outline-none focus:border-[#333] focus:ring-1 focus:ring-white/10 transition-all cursor-none"
            >
              <option value="" className="bg-[#111]">Select category</option>
              {categories?.map((c) => <option key={c.id} value={c.id} className="bg-[#111]">{c.icon} {c.name}</option>)}
            </select>
          </div>
        </div>

        <Input label="Location (optional)" placeholder="City, Country" {...register('location')} />
        <Input label="Tags (comma separated)" placeholder="design, logo, branding" {...register('tags')} />

        <div className="flex flex-col gap-2">
          <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Images (max 5)</label>
          <div className="flex flex-wrap gap-2">
            {previews.map((src, i) => (
              <div key={i} className="relative w-20 h-20 rounded-xl overflow-hidden border border-[#222]">
                <img src={src} alt="" className="w-full h-full object-cover" />
                <button type="button" onClick={() => removeImage(i)} className="absolute top-1 right-1 bg-black/70 text-white rounded-full p-0.5 hover:bg-black cursor-none">
                  <X size={10} />
                </button>
              </div>
            ))}
            {images.length < 5 && (
              <button
                type="button"
                onClick={() => fileRef.current?.click()}
                className="w-20 h-20 rounded-xl border border-dashed border-[#333] flex flex-col items-center justify-center text-zinc-600 hover:border-zinc-500 hover:text-zinc-400 transition-colors cursor-none"
              >
                <Upload size={16} />
                <span className="text-xs mt-1">Add</span>
              </button>
            )}
          </div>
          <input ref={fileRef} type="file" accept="image/*" multiple className="hidden" onChange={(e) => addImage(e.target.files)} />
        </div>

        <Button type="submit" size="lg" className="w-full" loading={isSubmitting}>
          Publish Listing
        </Button>
      </form>
    </div>
  )
}
