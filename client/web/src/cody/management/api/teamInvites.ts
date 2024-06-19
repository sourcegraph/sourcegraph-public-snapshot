import type { TeamRole } from './teamMembers'

export type TeamInviteStatus = 'sent' | 'errored' | 'accepted' | 'canceled'

export interface TeamInvite {
    id: string

    email: string
    role: TeamRole

    status: TeamInviteStatus
    error?: string

    sentAt: string
    sentBy: string
    acceptedAt?: string
}

export interface CreateTeamInviteRequest {
    email: string
    role: TeamRole
}

export interface ListTeamInvitesResponse {
    invites: Omit<TeamInvite, 'sentBy'>[]
    continuationToken?: string
}
