'use client'

import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { api } from '@/lib/api'
import DatePicker from '@/components/DatePicker'

export default function MyReservationsPage() {
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [limit] = useState(20)
  const [sortBy, setSortBy] = useState('StartAt')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [filters, setFilters] = useState<Record<string, string>>({})

  const { data: listMyReservationsData, isLoading: listMyReservationsDataLoading, error: listMyReservationsDataError } = useQuery({
    queryKey: ['ListMyReservations', page, limit, sortBy, sortDir, filters],
    queryFn: () => api.ListMyReservations({ page, limit, sortBy, sortDir, ...filters }),
  })

  const createReservationForm = useForm()
  const createReservationMutation = useMutation({
    mutationFn: (data: any) => api.CreateReservation(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ListMyReservations'] })
    },
  })

  return (
    <main className="max-w-4xl mx-auto p-6">
      <header className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">내 예약</h1>
      </header>
      {listMyReservationsDataLoading && <div>로딩 중...</div>}
      {listMyReservationsDataError && <div>오류가 발생했습니다</div>}
      {listMyReservationsData && (
        <section className="mb-8">
          <div className="flex gap-2 mb-4">
            <input placeholder="Status" value={filters.Status ?? ''} className="px-3 py-2 border rounded" onChange={(e) => setFilters(f => ({ ...f, Status: e.target.value }))} />
          </div>
          <div className="flex gap-2 mb-4">
            <button onClick={() => { setSortBy('StartAt'); setSortDir(d => d === 'asc' ? 'desc' : 'asc') }}>
              StartAt {sortBy === 'StartAt' ? (sortDir === 'asc' ? '↑' : '↓') : ''}
            </button>
          </div>
          <ul className="space-y-3">
            {listMyReservationsData.reservations?.map((item: any, index: number) => (
              <li key={index} className="flex justify-between items-center p-4 border rounded">
                <div>
                  <span className="font-semibold">{item.RoomID}</span>
                  <span className="ml-2 text-sm text-gray-500">{item.StartAt}</span>
                  <span className="text-sm text-gray-400">~</span>
                  <span className="text-sm text-gray-500">{item.EndAt}</span>
                </div>
                <span className="px-2 py-1 text-sm rounded bg-gray-100">{item.Status}</span>
              </li>
            ))}
          </ul>
          {listMyReservationsData.reservations?.length === 0 && <p className="text-gray-400 text-center py-8">예약이 없습니다</p>}
          <div className="flex justify-between items-center mt-4">
            <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}>이전</button>
            <span>{page} / {Math.ceil((listMyReservationsData?.total ?? 0) / limit)}</span>
            <button disabled={!listMyReservationsData?.total || page * limit >= listMyReservationsData.total} onClick={() => setPage(p => p + 1)}>다음</button>
          </div>
        </section>
      )}
      <section className="p-6 bg-gray-50 rounded-lg">
        <h2 className="text-lg font-semibold mb-4">새 예약</h2>
        <form onSubmit={createReservationForm.handleSubmit((data) => createReservationMutation.mutate(data))} className="space-y-4">
          <input type="number" placeholder="스터디룸 번호" className="w-full px-3 py-2 border rounded" {...createReservationForm.register('RoomID', { valueAsNumber: true })} />
          <div className="grid grid-cols-2 gap-4">
            <DatePicker {...createReservationForm.register('StartAt')} />
            <DatePicker {...createReservationForm.register('EndAt')} />
          </div>
          <button type="submit">예약하기</button>
        </form>
      </section>
    </main>
  )
}
