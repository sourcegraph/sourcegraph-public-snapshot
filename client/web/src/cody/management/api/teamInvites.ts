import { TeamRole } from './teamMembers'

export type TeamInviteStatus = 'sent' | 'errored' | 'accepted' | 'canceled'

export interface TeamInvite {
    id: string

    email: string
    role: TeamRole

    status: TeamInviteStatus
    error?: string

    sentAt: Date
    acceptedAt?: Date
}

export interface CreateTeamInviteRequest {
    email: string
    role: TeamRole
}

export interface ListTeamInvitesResponse {
    invites: TeamInvite[]
    continuationToken?: string
}
