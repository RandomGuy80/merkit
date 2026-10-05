'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { useAuth } from '@/lib/hooks/useAuth'
import { useToast } from '@/components/ui/Toast'
import { isAxiosError } from 'axios'

const schema = z.object({
  email: z.string().email('Invalid email'),
  password: z.string().min(1, 'Password required'),
})
type FormData = z.infer<typeof schema>

export function LoginForm() {
  const { login } = useAuth()
  const { toast } = useToast()
  const router = useRouter()
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const onSubmit = async (data: FormData) => {
    try {
      await login(data.email, data.password)
      toast('Welcome back!', 'success')
      router.push('/')
    } catch (err) {
      const msg = isAxiosError(err) ? err.response?.data?.error ?? 'Login failed' : 'Login failed'
      toast(msg, 'error')
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
      <Input label="Email" type="email" placeholder="you@example.com" error={errors.email?.message} {...register('email')} />
      <Input label="Password" type="password" placeholder="••••••••" error={errors.password?.message} {...register('password')} />
      <Button type="submit" className="w-full" size="lg" loading={isSubmitting}>
        Sign in
      </Button>
      <p className="text-center text-sm text-zinc-500">
        No account?{' '}
        <Link href="/register" className="text-zinc-200 hover:text-white font-medium transition-colors">Sign up</Link>
      </p>
    </form>
  )
}
