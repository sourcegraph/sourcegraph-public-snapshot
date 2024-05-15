import { useSSCQuery } from '../util'

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

const transformResponse = (response: MemberResponse): TeamMember[] => response.members

export const useCodyTeamMembers = (): [TeamMember[] | null, Error | null] =>
    useSSCQuery<MemberResponse, TeamMember[]>('/team/current/members', transformResponse)
