'use client'

import { use } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useEffect, useRef, useState } from 'react'
import { Upload, X, ArrowLeft } from 'lucide-react'
import Image from 'next/image'
import { listingsApi } from '@/lib/api/listings'
import { categoriesApi } from '@/lib/api/users'
import { useAuthStore } from '@/lib/store/auth'
import { Input, Textarea } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { Skeleton } from '@/components/ui/Skeleton'
import { useToast } from '@/components/ui/Toast'
import { getImageUrl } from '@/lib/utils'
import { isAxiosError } from 'axios'

const schema = z.object({
  title: z.string().min(3, 'Min 3 characters'),
  description: z.string().min(10, 'Min 10 characters'),
  price: z.string().min(1, 'Required'),
  category_id: z.string().optional(),
  location: z.string().optional(),
  tags: z.string().optional(),
  status: z.enum(['active', 'archived']),
})
type FormData = z.infer<typeof schema>

export default function EditListingPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const { user } = useAuthStore()
  const { toast } = useToast()
  const router = useRouter()
  const qc = useQueryClient()
  const fileRef = useRef<HTMLInputElement>(null)
  const [newFiles, setNewFiles] = useState<File[]>([])
  const [newPreviews, setNewPreviews] = useState<string[]>([])
  const [removingIdx, setRemovingIdx] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)

  const { data: listing, isLoading } = useQuery({
    queryKey: ['listing', id],
    queryFn: () => listingsApi.getById(id),
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: categoriesApi.list,
  })

  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  useEffect(() => {
    if (!listing) return
    reset({
      title: listing.title,
      description: listing.description,
      price: String(listing.price),
      category_id: listing.category_id ? String(listing.category_id) : '',
      location: listing.location ?? '',
      tags: listing.tags?.join(', ') ?? '',
      status: listing.status === 'archived' ? 'archived' : 'active',
    })
  }, [listing, reset])

  if (isLoading) return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-4">
      <Skeleton className="h-8 w-1/3" />
      <Skeleton className="h-64" />
    </div>
  )

  if (!listing) return <div className="text-center py-24 text-zinc-600">Listing not found</div>

  const isOwner = user?.id === listing.seller_id
  const isAdmin = user?.role === 'admin'
  if (!isOwner && !isAdmin) {
    router.push(`/listings/${id}`)
    return null
  }

  const addFiles = (files: FileList | null) => {
    if (!files) return
    const remaining = 5 - (listing.images?.length ?? 0) - newFiles.length
    const selected = Array.from(files).slice(0, remaining)
    const tooBig = selected.filter((f) => f.size > 10 * 1024 * 1024)
    if (tooBig.length) { toast('Images must be under 10MB', 'error'); return }
    setNewFiles((p) => [...p, ...selected])
    selected.forEach((f) => {
      const reader = new FileReader()
      reader.onload = (e) => setNewPreviews((p) => [...p, e.target?.result as string])
      reader.readAsDataURL(f)
    })
  }

  const removeNewFile = (i: number) => {
    setNewFiles((p) => p.filter((_, idx) => idx !== i))
    setNewPreviews((p) => p.filter((_, idx) => idx !== i))
  }

  const removeExisting = async (index: number) => {
    setRemovingIdx(index)
    try {
      await listingsApi.removeImage(id, index)
      qc.invalidateQueries({ queryKey: ['listing', id] })
      toast('Image removed', 'success')
    } catch {
      toast('Failed to remove image', 'error')
    } finally {
      setRemovingIdx(null)
    }
  }

  const onSubmit = async (data: FormData) => {
    const price = parseFloat(data.price)
    if (isNaN(price) || price <= 0) { toast('Price must be positive', 'error'); return }
    setSaving(true)
    try {
      await listingsApi.update(id, {
        title: data.title,
        description: data.description,
        price,
        category_id: data.category_id ? parseInt(data.category_id, 10) : undefined,
        location: data.location || undefined,
        tags: data.tags ? data.tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
        status: data.status,
      })
      for (const file of newFiles) {
        await listingsApi.addImage(id, file)
      }
      qc.invalidateQueries({ queryKey: ['listing', id] })
      toast('Listing updated!', 'success')
      router.push(`/listings/${id}`)
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Failed' : 'Failed'
      toast(msg, 'error')
    } finally {
      setSaving(false)
    }
  }

  const totalImages = (listing.images?.length ?? 0) + newFiles.length

  return (
    <div className="max-w-2xl mx-auto px-4 py-10">
      <button
        onClick={() => router.back()}
        className="flex items-center gap-1.5 text-xs text-zinc-600 hover:text-zinc-300 transition-colors mb-7 cursor-none"
      >
        <ArrowLeft size={14} /> Back
      </button>

      <h1 className="text-xl font-bold text-white tracking-tight mb-7">Edit Listing</h1>

      <form onSubmit={handleSubmit(onSubmit)} className="bg-[#111] rounded-2xl border border-[#222] p-6 space-y-5">
        <Input label="Title" error={errors.title?.message} {...register('title')} />
        <Textarea label="Description" rows={4} error={errors.description?.message} {...register('description')} />

        <div className="grid grid-cols-2 gap-4">
          <Input label="Price (USD)" type="number" step="0.01" error={errors.price?.message} {...register('price')} />
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Category</label>
            <select
              {...register('category_id')}
              className="w-full rounded-xl border border-[#222] px-4 py-2.5 text-sm bg-[#111] text-white focus:outline-none focus:border-[#333] focus:ring-1 focus:ring-white/10 transition-all cursor-none"
            >
              <option value="" className="bg-[#111]">Select category</option>
              {categories?.map((c) => (
                <option key={c.id} value={c.id} className="bg-[#111]">{c.icon} {c.name}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <Input label="Location (optional)" {...register('location')} />
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">Status</label>
            <select
              {...register('status')}
              className="w-full rounded-xl border border-[#222] px-4 py-2.5 text-sm bg-[#111] text-white focus:outline-none focus:border-[#333] focus:ring-1 focus:ring-white/10 transition-all cursor-none"
            >
              <option value="active" className="bg-[#111]">Active</option>
              <option value="archived" className="bg-[#111]">Archived</option>
            </select>
          </div>
        </div>

        <Input label="Tags (comma separated)" {...register('tags')} />

        {/* Images */}
        <div className="flex flex-col gap-2">
          <label className="text-xs font-medium text-zinc-400 uppercase tracking-wider">
            Images ({totalImages}/5)
          </label>
          <div className="flex flex-wrap gap-2">
            {/* Existing images */}
            {listing.images?.map((img, i) => (
              <div key={img} className="relative w-20 h-20 rounded-xl overflow-hidden border border-[#222]">
                <Image
                  src={getImageUrl(img)}
                  alt=""
                  fill
                  className="object-cover"
                  sizes="80px"
                  unoptimized={!img.startsWith('http')}
                />
                <button
                  type="button"
                  onClick={() => removeExisting(i)}
                  disabled={removingIdx === i}
                  className="absolute top-1 right-1 bg-black/70 text-white rounded-full p-0.5 hover:bg-red-600 transition-colors cursor-none disabled:opacity-50"
                >
                  <X size={10} />
                </button>
              </div>
            ))}

            {/* New images */}
            {newPreviews.map((src, i) => (
              <div key={i} className="relative w-20 h-20 rounded-xl overflow-hidden border border-[#333]">
                <img src={src} alt="" className="w-full h-full object-cover" />
                <button
                  type="button"
                  onClick={() => removeNewFile(i)}
                  className="absolute top-1 right-1 bg-black/70 text-white rounded-full p-0.5 hover:bg-black cursor-none"
                >
                  <X size={10} />
                </button>
                <div className="absolute bottom-1 left-1 text-[9px] bg-black/70 text-zinc-400 px-1 rounded">new</div>
              </div>
            ))}

            {totalImages < 5 && (
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
          <input ref={fileRef} type="file" accept="image/*" multiple className="hidden" onChange={(e) => addFiles(e.target.files)} />
        </div>

        <div className="flex gap-3 pt-1">
          <Button type="button" variant="outline" onClick={() => router.back()}>
            Cancel
          </Button>
          <Button type="submit" size="lg" className="flex-1" loading={saving}>
            Save changes
          </Button>
        </div>
      </form>
    </div>
  )
}
