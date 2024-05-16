import { useSSCQuery } from '../util'

type TeamRole = 'member' | 'admin'

interface CodySubscriptionSummary {
    teamId: string
    userRole: TeamRole
}

export const useCodySubscriptionSummaryData = (): [CodySubscriptionSummary | null, Error | null] =>
    useSSCQuery<CodySubscriptionSummary>('/team/current/subscription/summary')
