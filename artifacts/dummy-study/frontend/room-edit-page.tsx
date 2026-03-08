'use client'

import React from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { api } from '@/lib/api'

export default function RoomEditPage() {
  const { RoomID } = useParams()
  const queryClient = useQueryClient()

  const updateRoomForm = useForm()
  const updateRoomMutation = useMutation({
    mutationFn: (data: any) => api.UpdateRoom({ ...data, roomId: RoomID }),
    onSuccess: () => {
      queryClient.invalidateQueries()
    },
  })

  const deleteRoomMutation = useMutation({
    mutationFn: (data: any) => api.DeleteRoom({ ...data, roomId: RoomID }),
    onSuccess: () => {
      queryClient.invalidateQueries()
    },
  })

  return (
    <main className="max-w-2xl mx-auto p-6">
      <header className="mb-6">
        <h1 className="text-2xl font-bold">스터디룸 수정</h1>
      </header>
      <form onSubmit={updateRoomForm.handleSubmit((data) => updateRoomMutation.mutate(data))} className="space-y-4">
        <input placeholder="스터디룸 이름" className="w-full px-3 py-2 border rounded" {...updateRoomForm.register('Name')} />
        <input type="number" placeholder="수용 인원" className="w-full px-3 py-2 border rounded" {...updateRoomForm.register('Capacity', { valueAsNumber: true })} />
        <input placeholder="위치" className="w-full px-3 py-2 border rounded" {...updateRoomForm.register('Location')} />
        <button type="submit">수정</button>
      </form>
      <footer className="mt-8 pt-4 border-t">
        <button onClick={() => deleteRoomMutation.mutate({})} className="w-full py-2 bg-red-500 text-white rounded">스터디룸 삭제</button>
      </footer>
    </main>
  )
}
