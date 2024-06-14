export type TeamRole = 'member' | 'admin'

export interface TeamMember {
    accountId: string
    displayName: string
    email: string
    avatarUrl: string
    role: TeamRole
}

export interface TeamMemberRef {
    accountId: string
    teamRole: TeamRole
}

export interface ListTeamMembersResponse {
    members: TeamMember[]
    continuationToken?: string
}

export interface UpdateTeamMembersRequest {
    addMember?: TeamMemberRef
    removeMember?: TeamMemberRef
    updateMemberRole?: TeamMemberRef
}
