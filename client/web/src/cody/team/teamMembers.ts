import { useCallback } from 'react'

import { useSSCData } from '../util'

export interface TeamMember {
    accountId: string
    displayName: string | null
    email: string
    avatarUrl: string | null
    role: 'admin' | 'member'
}

interface MemberResponse {
    members: TeamMember[]
}

export const useCodyTeamMembers = (): [TeamMember[] | null, Error | null] => {
    return useSSCData<MemberResponse, TeamMember[]>(
        '/team/current/members',
        useCallback(response => response.members, [])
    )
}
