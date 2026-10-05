import { LoginForm } from '@/components/auth/LoginForm'

export const metadata = { title: 'Sign in — Merkit' }

export default function LoginPage() {
  return (
    <div className="min-h-[calc(100vh-8rem)] flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-white tracking-tight">Welcome back</h1>
          <p className="text-zinc-500 mt-1.5 text-sm">Sign in to your Merkit account</p>
        </div>
        <div className="bg-[#111] rounded-2xl border border-[#222] p-8">
          <LoginForm />
        </div>
      </div>
    </div>
  )
}
