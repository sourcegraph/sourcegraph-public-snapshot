import { useCallback } from 'react'

import { useSSCData } from '../util'

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

export const useCodyTeamInvites = (): [TeamInvite[] | null, Error | null] => {
    return useSSCData<InviteResponse, TeamInvite[]>(
        '/team/current/invites',
        useCallback(response => response.invites, [])
    )
}
