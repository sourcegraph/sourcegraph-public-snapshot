import { useEffect, useMemo, useState } from 'react'

import { CodyProApiError } from '../management/api/react-query/callCodyProApi'
import { useInvite } from '../management/api/react-query/invites'
import { useSubscriptionSummary } from '../management/api/react-query/subscriptions'
import { useTeamMembers } from '../management/api/react-query/teams'
import type { TeamInvite } from '../management/api/teamInvites'

export enum UserInviteStatus {
    Error = 'error',

    NoCurrentTeam = 'no_current_team',
    InvitedTeamMember = 'invited_team_member',
    AnotherTeamMember = 'another_team_member',
    AnotherTeamSoleAdmin = 'another_team_sole_admin',
}

type UseInviteStateHook = (
    teamId: string,
    inviteId: string
) =>
    | { status: 'loading' }
    | { status: 'error' }
    | {
          status: 'success'
          initialInviteStatus: TeamInvite['status']
          sentBy: TeamInvite['sentBy']
          initialUserStatus: UserInviteStatus
      }

export const useInviteState: UseInviteStateHook = (teamId, inviteId) => {
    const inviteQuery = useInvite({ teamId, inviteId })
    const subscriptionSummaryQuery = useSubscriptionSummary()
    const teamMembersQuery = useTeamMembers()

    const [initialInviteStatus, setInitialInviteStatus] = useState<TeamInvite['status']>()
    const [initialUserStatus, setInitialUserStatus] = useState<UserInviteStatus>()

    useEffect(() => {
        setInitialUserStatus(prevStatus => {
            // If user status is already defined, use it.
            if (prevStatus !== undefined) {
                return prevStatus
            }

            // There are two distinct cases when subscription summary query may fail:
            // 1. 404 error indicating that user is not on a team yet. Return the no current team status.
            // 2. Other kind of error indicating that we failed to get subscription data for other reason.
            // We can't define user status without subscription data so retunr the error status.
            if (subscriptionSummaryQuery.isError) {
                return subscriptionSummaryQuery.error instanceof CodyProApiError &&
                    subscriptionSummaryQuery.error.status === 404
                    ? UserInviteStatus.NoCurrentTeam
                    : UserInviteStatus.Error
            }

            // Wait for the subscription summary query to succeed. We handle the error case above.
            if (!subscriptionSummaryQuery.data) {
                return undefined
            }

            // Team members query is executed only if the user is an admin.
            // If it fails, we can't define whether the user is the sole admin of the team.
            // Return the error status.
            if (teamMembersQuery.isError) {
                return UserInviteStatus.Error
            }

            // Now subscription summary is available.

            // User is already on the team they have been invited to.
            if (subscriptionSummaryQuery.data.teamId === teamId) {
                return UserInviteStatus.InvitedTeamMember
            }

            // User is on another team.

            // If user is admin, check if they are a sole admin on a team.
            if (subscriptionSummaryQuery.data.userRole === 'admin') {
                if (!teamMembersQuery.isSuccess) {
                    // Waiting for team members query to succeed. We handle the error case above.
                    return undefined
                }

                if (!teamMembersQuery.data) {
                    // Team members query is fetched, but data is undefined.
                    // We can define whether the user is a sole admin on a team.
                    // Return error status.
                    return UserInviteStatus.Error
                }

                const currentTeamAdminsCount = teamMembersQuery.data.members.filter(
                    member => member.role === 'admin'
                ).length
                if (currentTeamAdminsCount === 1) {
                    return UserInviteStatus.AnotherTeamSoleAdmin
                }
            }

            // User is either a member or one of several admins (not the sole admin) of another team.
            return UserInviteStatus.AnotherTeamMember
        })
    }, [teamId, subscriptionSummaryQuery, teamMembersQuery, setInitialUserStatus])

    useEffect(() => {
        setInitialInviteStatus(s => s || inviteQuery.data?.status)
    }, [inviteQuery, setInitialInviteStatus])

    const state: ReturnType<UseInviteStateHook> = useMemo(() => {
        if (inviteQuery.isError || (inviteQuery.isSuccess && !inviteQuery.data)) {
            return { status: 'error' }
        }

        if (!inviteQuery.data || !initialInviteStatus || !initialUserStatus) {
            return { status: 'loading' }
        }

        return {
            status: 'success',
            initialInviteStatus,
            initialUserStatus,
            sentBy: inviteQuery.data.sentBy,
        }
    }, [inviteQuery.isError, inviteQuery.isSuccess, inviteQuery.data, initialInviteStatus, initialUserStatus])

    return state
}
