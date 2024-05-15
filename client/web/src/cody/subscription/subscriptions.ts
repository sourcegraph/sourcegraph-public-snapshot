import { useSSCQuery } from '../util'

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

const transformSubscriptionResponse = (response: SubscriptionResponse): SubscriptionData => ({
    seatCount: response.maxSeats,
    isPro: response.subscriptionStatus !== 'canceled',
})
export const useCodySubscriptionData = (): [SubscriptionData | null, Error | null] =>
    useSSCQuery<SubscriptionResponse, SubscriptionData>('/team/current/subscription', transformSubscriptionResponse)

const transformSummaryResponse = (response: SubscriptionSummaryResponse): SubscriptionSummaryData => ({
    teamId: response.teamId,
    isAdmin: response.userRole === 'admin',
})
export const useCodySubscriptionSummaryData = (): [SubscriptionSummaryData | null, Error | null] =>
    useSSCQuery<SubscriptionSummaryResponse, SubscriptionSummaryData>(
        '/team/current/subscription/summary',
        transformSummaryResponse
    )
