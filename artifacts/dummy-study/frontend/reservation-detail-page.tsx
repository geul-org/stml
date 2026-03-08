'use client'

import React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { api } from '@/lib/api'

export default function ReservationDetailPage() {
  const { ReservationID } = useParams()
  const queryClient = useQueryClient()

  const { data: getReservationData, isLoading: getReservationDataLoading, error: getReservationDataError } = useQuery({
    queryKey: ['GetReservation', ReservationID],
    queryFn: () => api.GetReservation({ reservationId: ReservationID }),
  })

  const cancelReservationMutation = useMutation({
    mutationFn: (data: any) => api.CancelReservation({ ...data, reservationId: ReservationID }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['GetReservation'] })
    },
  })

  return (
    <main className="max-w-2xl mx-auto p-6">
      {getReservationDataLoading && <div>로딩 중...</div>}
      {getReservationDataError && <div>오류가 발생했습니다</div>}
      {getReservationData && (
        <article>
          <header className="flex justify-between items-center mb-6">
            <h1 className="text-2xl font-bold">예약 상세</h1>
            <span className="px-3 py-1 text-sm rounded bg-gray-100">{getReservationData.reservation.Status}</span>
          </header>
          <dl className="space-y-3">
            <div className="flex justify-between">
              <dt className="text-gray-500">스터디룸</dt>
              <dd className="font-semibold">{getReservationData.reservation.RoomID}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">시작</dt>
              <dd>{getReservationData.reservation.StartAt}</dd>
            </div>
            <div className="flex justify-between">
              <dt className="text-gray-500">종료</dt>
              <dd>{getReservationData.reservation.EndAt}</dd>
            </div>
          </dl>
          {getReservationData.canCancel && (
            <footer className="mt-8 pt-4 border-t">
              <button onClick={() => cancelReservationMutation.mutate({})} className="w-full py-2 bg-red-500 text-white rounded">예약 취소</button>
            </footer>
          )}
        </article>
      )}
    </main>
  )
}
