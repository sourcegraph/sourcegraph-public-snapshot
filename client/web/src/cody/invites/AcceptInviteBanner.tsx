import { Button, ButtonLink, H1, Text } from '@sourcegraph/wildcard'

import { CodyProRoutes } from '../codyProRoutes'
import { CodyAlert } from '../components/CodyAlert'
import { useAcceptInvite, useCancelInvite } from '../management/api/react-query/invites'

import { useInviteParams } from './useInviteParams'
import { UserInviteStatus, useInviteState } from './useInviteState'

export const AcceptInviteBanner: React.FC<{ onSuccess: () => unknown }> = ({ onSuccess }) => {
    const { inviteParams, clearInviteParams } = useInviteParams()
    if (!inviteParams) {
        return null
    }
    return (
        <AcceptInviteBannerContent
            teamId={inviteParams.teamId}
            inviteId={inviteParams.inviteId}
            onSuccess={onSuccess}
            clearInviteParams={clearInviteParams}
        />
    )
}

const AcceptInviteBannerContent: React.FC<{
    teamId: string
    inviteId: string
    onSuccess: () => unknown
    clearInviteParams: () => void
}> = ({ teamId, inviteId, onSuccess, clearInviteParams }) => {
    const inviteState = useInviteState(teamId, inviteId)
    const acceptInviteMutation = useAcceptInvite()
    const cancelInviteMutation = useCancelInvite()

    if (inviteState.status === 'loading') {
        return null
    }

    if (
        inviteState.status === 'error' ||
        inviteState.initialInviteStatus !== 'sent' ||
        inviteState.initialUserStatus === UserInviteStatus.Error
    ) {
        return (
            <CodyAlert variant="error">
                <H1 as="p" className="mb-2">
                    Issue with invite
                </H1>
                <Text className="mb-0">The invitation is no longer valid. Contact your team admin.</Text>
            </CodyAlert>
        )
    }

    switch (inviteState.initialUserStatus) {
        case UserInviteStatus.NoCurrentTeam:
        case UserInviteStatus.AnotherTeamMember: {
            // Invite has been canceled. Remove the banner.
            if (cancelInviteMutation.isSuccess || cancelInviteMutation.isError) {
                return null
            }

            switch (acceptInviteMutation.status) {
                case 'error': {
                    return (
                        <CodyAlert variant="error">
                            <H1 as="p" className="mb-2">
                                Issue with invite
                            </H1>
                            <Text className="mb-0">
                                Accepting invite failed with error: {acceptInviteMutation.error.message}.
                            </Text>
                        </CodyAlert>
                    )
                }
                case 'success': {
                    return (
                        <CodyAlert variant="greenCodyPro">
                            <H1 as="p" className="mb-2">
                                Pro team change complete!
                            </H1>
                            <Text>
                                {inviteState.initialUserStatus === UserInviteStatus.NoCurrentTeam
                                    ? 'You successfully joined the new Cody Pro team.'
                                    : 'Your pro team has been successfully changed.'}
                            </Text>
                        </CodyAlert>
                    )
                }
                case 'idle':
                case 'pending':
                default: {
                    return (
                        <CodyAlert variant="purple">
                            <H1 as="p" className="mb-2">
                                Join new Cody Pro team?
                            </H1>
                            <Text>You've been invited to a new Cody Pro team by {inviteState.sentBy}.</Text>
                            <Text>
                                {inviteState.initialUserStatus === UserInviteStatus.NoCurrentTeam
                                    ? 'You will get unlimited autocompletions, chat messages and commands.'
                                    : 'This will terminate your current Cody Pro plan, and place you on the new Cody Pro team. You will not lose access to your Cody Pro benefits.'}
                            </Text>
                            <div>
                                <Button
                                    variant="primary"
                                    disabled={acceptInviteMutation.isPending || cancelInviteMutation.isPending}
                                    className="mr-3"
                                    onClick={() =>
                                        acceptInviteMutation.mutate(
                                            { teamId, inviteId },
                                            { onSuccess, onSettled: clearInviteParams }
                                        )
                                    }
                                >
                                    Accept
                                </Button>
                                <Button
                                    variant="link"
                                    disabled={acceptInviteMutation.isPending || cancelInviteMutation.isPending}
                                    onClick={() =>
                                        cancelInviteMutation.mutate(
                                            { teamId, inviteId },
                                            { onSettled: clearInviteParams }
                                        )
                                    }
                                >
                                    Decline
                                </Button>
                            </div>
                        </CodyAlert>
                    )
                }
            }
        }
        case UserInviteStatus.InvitedTeamMember: {
            if (cancelInviteMutation.isIdle) {
                void cancelInviteMutation.mutate({ teamId, inviteId }, { onSettled: clearInviteParams })
            }
            return (
                <CodyAlert variant="error">
                    <H1 as="p" className="mb-2">
                        Issue with invite
                    </H1>
                    <Text className="mb-0">
                        You've been invited to a Cody Pro team by {inviteState.sentBy}.<br />
                        You cannot accept this invite as as you are already on this team.
                    </Text>
                </CodyAlert>
            )
        }
        case UserInviteStatus.AnotherTeamSoleAdmin: {
            return (
                <CodyAlert variant="error">
                    <H1 as="p" className="mb-2">
                        Issue with invite
                    </H1>
                    <Text className="mb-0">You've been invited to a new Cody Pro team by {inviteState.sentBy}.</Text>
                    <Text>
                        To accept this invite you need to transfer your administrative role to another member of your
                        team and click the invite link again.
                    </Text>
                    <div>
                        <ButtonLink variant="primary" to={CodyProRoutes.ManageTeam}>
                            Manage team
                        </ButtonLink>
                    </div>
                </CodyAlert>
            )
        }
        default: {
            return null
        }
    }
}
