import { useCallback } from 'react'

import { useSSCData } from '../util'

interface SubscriptionResponse {
    subscriptionStatus: 'active' | 'past_due' | 'unpaid' | 'canceled' | 'trialing' | 'other'
    maxSeats: number
}

interface SubscriptionSummaryResponse {
    teamId: string
    userRole: 'none' | 'member' | 'admin'
}

interface SubscriptionData {
    seatCount: number | null
    isPro: boolean | null
}

interface SubscriptionSummaryData {
    teamId: string | null
    isAdmin: boolean | null
}

export const useCodySubscriptionData = (): [SubscriptionData | null, Error | null] => {
    return useSSCData<SubscriptionResponse, SubscriptionData>(
        '/team/current/subscription',
        useCallback(
            response => ({ seatCount: response.maxSeats, isPro: response.subscriptionStatus !== 'canceled' }),
            []
        )
    )
}

export const useCodySubscriptionSummaryData = (): [SubscriptionSummaryData | null, Error | null] => {
    return useSSCData<SubscriptionSummaryResponse, SubscriptionSummaryData>(
        '/team/current/subscription/summary',
        useCallback(response => ({ teamId: response.teamId, isAdmin: response.userRole === 'admin' }), [])
    )
}
