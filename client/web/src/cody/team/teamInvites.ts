import { useSSCQuery } from '../util'

export interface TeamInvite {
    id: string
    email: string
    role: 'admin' | 'member' | 'none'
    status: 'sent' | 'errored' | 'accepted' | 'canceled'
    error: string | null
    sentAt: string | null
    acceptedAt: string | null
}

interface InviteResponse {
    invites: TeamInvite[]
}

const transformResponse = (response: InviteResponse): TeamInvite[] => response.invites

export const useCodyTeamInvites = (): [TeamInvite[] | null, Error | null] =>
    useSSCQuery<InviteResponse, TeamInvite[]>('/team/current/invites', transformResponse)
