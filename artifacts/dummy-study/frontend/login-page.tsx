'use client'

import React from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { api } from '@/lib/api'

export default function LoginPage() {
  const queryClient = useQueryClient()

  const loginForm = useForm()
  const loginMutation = useMutation({
    mutationFn: (data: any) => api.Login(data),
    onSuccess: () => {
      queryClient.invalidateQueries()
    },
  })

  return (
    <main className="flex items-center justify-center min-h-screen bg-gray-50">
      <div className="w-full max-w-sm p-6 bg-white rounded-lg shadow">
        <h1 className="text-2xl font-bold mb-6 text-center">로그인</h1>
        <form onSubmit={loginForm.handleSubmit((data) => loginMutation.mutate(data))} className="space-y-4">
          <input type="email" placeholder="이메일" className="w-full px-3 py-2 border rounded" {...loginForm.register('Email')} />
          <input type="password" placeholder="비밀번호" className="w-full px-3 py-2 border rounded" {...loginForm.register('Password')} />
          <button type="submit">로그인</button>
        </form>
      </div>
    </main>
  )
}
