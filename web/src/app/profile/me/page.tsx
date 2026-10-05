'use client'

import { useQueryClient } from '@tanstack/react-query'
import { useRef, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { Camera } from 'lucide-react'
import { usersApi } from '@/lib/api/users'
import { useAuthStore } from '@/lib/store/auth'
import { Avatar } from '@/components/ui/Avatar'
import { Button } from '@/components/ui/Button'
import { Input, Textarea } from '@/components/ui/Input'
import { Stars } from '@/components/ui/Stars'
import { Skeleton } from '@/components/ui/Skeleton'
import { useToast } from '@/components/ui/Toast'
import { Badge } from '@/components/ui/Badge'
import { isAxiosError } from 'axios'
import { useEffect } from 'react'

export default function MyProfilePage() {
  const { user, updateUser, hydrated } = useAuthStore()
  const { toast } = useToast()
  const router = useRouter()
  const qc = useQueryClient()
  const fileRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)

  useEffect(() => {
    if (hydrated && !user) router.push('/login')
  }, [hydrated, user, router])

  const { register, handleSubmit, formState: { isSubmitting } } = useForm({
    values: { name: user?.name ?? '', bio: user?.bio ?? '' },
  })

  const onSubmit = async (data: { name: string; bio: string }) => {
    try {
      const updated = await usersApi.updateMe({ name: data.name, bio: data.bio || undefined })
      updateUser(updated)
      toast('Profile updated', 'success')
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Failed' : 'Failed'
      toast(msg, 'error')
    }
  }

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (file.size > 10 * 1024 * 1024) { toast('Max 10MB', 'error'); return }
    setUploading(true)
    try {
      await usersApi.uploadAvatar(file)
      const updated = await usersApi.getMe()
      // append timestamp to bust browser HTTP cache when same filename is reused
      if (updated.avatar) {
        const base = updated.avatar.split('?')[0]
        updated.avatar = `${base}?t=${Date.now()}`
      }
      updateUser(updated)
      qc.invalidateQueries({ queryKey: ['me'] })
      toast('Avatar updated', 'success')
    } catch {
      toast('Failed to upload avatar', 'error')
    } finally {
      setUploading(false)
    }
  }

  if (!user) return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-4">
      <Skeleton className="h-32" /><Skeleton className="h-48" />
    </div>
  )

  return (
    <div className="max-w-2xl mx-auto px-4 py-10 space-y-4">
      <h1 className="text-xl font-bold text-white tracking-tight">My Profile</h1>

      <div className="bg-[#111] rounded-2xl border border-[#222] p-6 flex items-center gap-6">
        <div className="relative">
          <Avatar src={user.avatar} name={user.name} size={80} />
          <button
            onClick={() => fileRef.current?.click()}
            disabled={uploading}
            className="absolute -bottom-1 -right-1 w-7 h-7 bg-white text-black rounded-full flex items-center justify-center hover:bg-zinc-200 transition-colors disabled:opacity-50 cursor-none"
          >
            <Camera size={13} />
          </button>
          <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={handleAvatarUpload} />
        </div>
        <div className="flex-1">
          <p className="font-bold text-xl text-white">{user.name}</p>
          <p className="text-zinc-500 text-sm">{user.email}</p>
          <div className="flex items-center gap-3 mt-2">
            <Badge variant={user.role === 'seller' ? 'purple' : user.role === 'admin' ? 'danger' : 'info'}>
              {user.role}
            </Badge>
            {user.role === 'seller' && (
              <div className="flex items-center gap-1.5">
                <Stars rating={user.rating} size={12} />
                <span className="text-xs text-zinc-600">{user.review_count} reviews</span>
              </div>
            )}
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="bg-[#111] rounded-2xl border border-[#222] p-6 space-y-4">
        <h2 className="text-xs font-medium text-zinc-500 uppercase tracking-wider">Edit Profile</h2>
        <Input label="Full name" {...register('name')} />
        <Textarea label="Bio" rows={3} placeholder="Tell buyers about yourself..." {...register('bio')} />
        <Button type="submit" loading={isSubmitting}>Save changes</Button>
      </form>
    </div>
  )
}
