import { RegisterForm } from '@/components/auth/RegisterForm'

export const metadata = { title: 'Sign up — Merkit' }

export default function RegisterPage() {
  return (
    <div className="min-h-[calc(100vh-8rem)] flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-white tracking-tight">Create account</h1>
          <p className="text-zinc-500 mt-1.5 text-sm">Join Merkit to buy or sell</p>
        </div>
        <div className="bg-[#111] rounded-2xl border border-[#222] p-8">
          <RegisterForm />
        </div>
      </div>
    </div>
  )
}
