import React, { useEffect, useMemo, useState } from 'react'

import type { MutateOptions } from '@tanstack/react-query'
import { Navigate, useLocation, useSearchParams } from 'react-router-dom'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { CodyProRoutes } from '../codyProRoutes'
import { CodyProApiError } from '../management/api/react-query/callCodyProApi'
import { useAcceptInvite, useCancelInvite, useInvite } from '../management/api/react-query/invites'
import { useSubscriptionSummary } from '../management/api/react-query/subscriptions'
import { useTeamMembers } from '../management/api/react-query/teams'
import type { TeamInvite } from '../management/api/teamInvites'

interface CodyAcceptInvitePageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedCodyAcceptInvitePage: React.FunctionComponent<CodyAcceptInvitePageProps> = ({
    telemetryRecorder,
}) => {
    const location = useLocation()
    const inviteState = useInviteState()

    useEffect(() => {
        telemetryRecorder.recordEvent('cody.invites.accept', 'view')
    }, [telemetryRecorder])

    switch (inviteState.status) {
        case 'pending': {
            return null
        }
        case 'error': {
            return <Navigate to={CodyProRoutes.Manage} replace={true} />
        }
        case 'success': {
            return <Navigate to={CodyProRoutes.Manage + location.search} replace={true} state={{ fromInvite: true }} />
        }
        default: {
            throw new Error('Unexpected invite state status')
        }
    }
}

export const CodyAcceptInvitePage = withAuthenticatedUser(AuthenticatedCodyAcceptInvitePage)

const useInviteParams = (): { params?: { teamId: string; inviteId: string }; clear: () => void } => {
    const [params, setParams] = useSearchParams()
    const teamId = params.get('teamID')
    const inviteId = params.get('inviteID')

    return {
        params: teamId && inviteId ? { teamId, inviteId } : undefined,
        clear: () =>
            setParams(
                prev => {
                    prev.delete('teamID')
                    prev.delete('inviteID')
                    return prev
                },
                { replace: true }
            ),
    }
}

export enum UserInviteStatus {
    Error,

    NoCurrentTeam,
    InvitedTeamMember,
    AnotherTeamMember,
    AnotherTeamSoleAdmin,
}

type InviteMutateFn = (
    options?: Pick<
        MutateOptions<unknown, Error, { teamId: string; inviteId: string }, unknown>,
        'onSuccess' | 'onError' | 'onSettled'
    >
) => void

type UseInviteStateHook = () =>
    | { status: 'pending' }
    | { status: 'error' }
    | {
          status: 'success'
          invite: TeamInvite
          userStatus: UserInviteStatus
          acceptInviteMutation: ReturnType<typeof useAcceptInvite> & { mutate: InviteMutateFn }
          cancelInviteMutation: ReturnType<typeof useCancelInvite> & { mutate: InviteMutateFn }
      }

export const useInviteState: UseInviteStateHook = () => {
    const { params: inviteParams } = useInviteParams()
    const inviteQuery = useInvite(inviteParams)
    const subscriptionSummaryQuery = useSubscriptionSummary()
    const teamMembersQuery = useTeamMembers()
    const acceptInviteMutation = useAcceptInvite()
    const cancelInviteMutation = useCancelInvite()

    const [userStatus, setUserStatus] = useState<UserInviteStatus>()

    useEffect(() => {
        setUserStatus(prevStatus => {
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

            // Team members query is executed only if the user is an admin.
            // If it fails, we can't define whether the user is the sole admin of the team.
            // Return the error status.
            if (teamMembersQuery.isError) {
                return UserInviteStatus.Error
            }

            // Wait for the subscription summary query to succeed. We handle the error case above.
            if (!subscriptionSummaryQuery.data) {
                return undefined
            }

            // Now subscription summary is available.

            // User is already on the team they have been invited to.
            if (subscriptionSummaryQuery.data.teamId === inviteParams?.teamId) {
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
    }, [inviteParams?.teamId, subscriptionSummaryQuery, teamMembersQuery])

    const state = useMemo(() => {
        if (!inviteQuery.data) {
            if (inviteParams && inviteQuery.isError) {
                return { status: 'error' as const }
            }

            return { status: 'pending' as const }
        }

        // Invite data is available.

        // Wait for user status to be defined.
        if (userStatus === undefined) {
            return { status: 'pending' as const }
        }

        // User status is defined.

        // TODO: consider clearing invite params after mutation succeeds
        // const clearInviteParams = (): void =>
        //     setParams(params => {
        //         params.delete('inviteID')
        //         params.delete('teamID')
        //         return params
        //     })

        const acceptInvite: InviteMutateFn = options =>
            inviteParams ? acceptInviteMutation.mutate(inviteParams, options) : undefined

        const cancelInvite: InviteMutateFn = options =>
            inviteParams ? cancelInviteMutation.mutate(inviteParams, options) : undefined

        return {
            status: 'success' as const,
            invite: inviteQuery.data,
            userStatus,
            acceptInviteMutation: { ...acceptInviteMutation, mutate: acceptInvite },
            cancelInviteMutation: { ...cancelInviteMutation, mutate: cancelInvite },
        }
    }, [acceptInviteMutation, cancelInviteMutation, inviteParams, inviteQuery.data, inviteQuery.isError, userStatus])

    return state
}
