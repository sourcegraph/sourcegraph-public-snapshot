import React, { useEffect } from 'react'

import { Navigate, useLocation, useSearchParams } from 'react-router-dom'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { CodyProRoutes } from '../codyProRoutes'
import { CodyProApiError } from '../management/api/react-query/callCodyProApi'
import { useSubscriptionSummary } from '../management/api/react-query/subscriptions'
import { useTeamMembers } from '../management/api/react-query/teams'

interface CodyAcceptInvitePageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedCodyAcceptInvitePage: React.FunctionComponent<CodyAcceptInvitePageProps> = ({
    telemetryRecorder,
}) => {
    const location = useLocation()
    const userInviteStatus = useUserInviteStatus()

    useEffect(() => {
        telemetryRecorder.recordEvent('cody.invites.accept', 'view')
    }, [telemetryRecorder])

    switch (userInviteStatus) {
        case undefined: {
            return null
        }
        case UserInviteStatus.AnotherTeamSoleAdmin: {
            return <Navigate to={CodyProRoutes.ManageTeam + location.search} replace={true} />
        }
        default: {
            return <Navigate to={CodyProRoutes.Manage + location.search} replace={true} />
        }
    }
}

export const CodyAcceptInvitePage = withAuthenticatedUser(AuthenticatedCodyAcceptInvitePage)

export const useInviteParams = (): { params?: { teamId: string; inviteId: string }; clear: () => void } => {
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
export const useUserInviteStatus = (): UserInviteStatus | undefined => {
    const { params: inviteParamsFromURL } = useInviteParams()
    const subscriptionSummaryQueryEnabled = inviteParamsFromURL !== undefined
    const subscriptionSummaryQuery = useSubscriptionSummary({ enabled: subscriptionSummaryQueryEnabled })
    const teamMembersQueryEnabled = subscriptionSummaryQuery.data?.userRole === 'admin'
    const teamMembersQuery = useTeamMembers({ enabled: teamMembersQueryEnabled })

    if (!inviteParamsFromURL) {
        return UserInviteStatus.Error
    }

    if (
        (subscriptionSummaryQueryEnabled ? !subscriptionSummaryQuery.isFetched : false) ||
        (teamMembersQueryEnabled ? !teamMembersQuery.isFetched : false)
    ) {
        return undefined
    }

    if (subscriptionSummaryQuery.isError) {
        return subscriptionSummaryQuery.error instanceof CodyProApiError &&
            subscriptionSummaryQuery.error.status === 404
            ? UserInviteStatus.NoCurrentTeam
            : UserInviteStatus.Error
    }

    if (teamMembersQuery.isError) {
        return UserInviteStatus.Error
    }

    if (subscriptionSummaryQuery.data && (teamMembersQueryEnabled ? teamMembersQuery.data : true)) {
        const { teamId, userRole } = subscriptionSummaryQuery.data

        if (teamId) {
            if (teamId === inviteParamsFromURL.teamId) {
                return UserInviteStatus.InvitedTeamMember
            }
            if (userRole === 'admin') {
                const currentTeamAdminsCount = teamMembersQuery.data!.members.filter(
                    member => member.role === 'admin'
                ).length
                if (currentTeamAdminsCount === 1) {
                    return UserInviteStatus.AnotherTeamSoleAdmin
                }
            }
            return UserInviteStatus.AnotherTeamMember
        }
    }

    return undefined
}
