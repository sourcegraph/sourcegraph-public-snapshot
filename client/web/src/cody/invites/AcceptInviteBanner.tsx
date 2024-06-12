import { useNavigate, useSearchParams } from 'react-router-dom'

import { Button, H3, Text } from '@sourcegraph/wildcard'

import { CodyProRoutes } from '../codyProRoutes'
import { CodyAlert } from '../components/CodyAlert'
import { useAcceptInvite, useCancelInvite } from '../management/api/react-query/invites'
import { useSubscriptionSummary } from '../management/api/react-query/subscriptions'

export const AcceptInviteBanner: React.FC = () => {
    const navigate = useNavigate()
    const [searchParams, setSearchParams] = useSearchParams()
    const teamIdFromURL = searchParams.get('teamID')
    const inviteIdFromURL = searchParams.get('inviteID')

    const subscriptionSummaryQuery = useSubscriptionSummary({ enabled: Boolean(teamIdFromURL && inviteIdFromURL) })
    const acceptInviteMutation = useAcceptInvite()
    const cancelInviteMutation = useCancelInvite()

    if (!teamIdFromURL || !inviteIdFromURL) {
        return null
    }

    if (subscriptionSummaryQuery.isLoading || subscriptionSummaryQuery.isError || !subscriptionSummaryQuery.data) {
        // TODO: consider provding UI feedback for these states
        return null
    }

    // user is on this team already
    if (subscriptionSummaryQuery.data.teamId === teamIdFromURL) {
        if (cancelInviteMutation.isIdle) {
            void cancelInviteMutation.mutate(
                { teamId: teamIdFromURL, inviteId: inviteIdFromURL },
                { onSuccess: () => setSearchParams({}, { replace: true }) }
            )
        }
        // TODO: handle cancel invite error state
        return (
            <CodyAlert variant="purple">
                <H3 className="mt-4">Can't accept an invite</H3>
                <Text>You are already memeber of the team you've beem invited to.</Text>
            </CodyAlert>
        )
    }

    // user is on another team
    if (subscriptionSummaryQuery.data.teamId) {
        return (
            <CodyAlert variant="purple">
                <H3 className="mt-4">Join new Cody Pro team?</H3>
                <Text>You've been invited to a new Cody Pro team by rob@acmecorp.com.</Text>
                <Text>
                    To accept this invite you need to transfer your administrative role to another member of your team.
                </Text>
                <div>
                    <Button
                        onClick={() =>
                            acceptInviteMutation.mutate(
                                { teamId: teamIdFromURL, inviteId: inviteIdFromURL },
                                { onSuccess: () => navigate(`${CodyProRoutes.Manage}?welcome=1`, { replace: true }) }
                            )
                        }
                    >
                        Accept
                    </Button>
                    <Button variant="secondary">Decline</Button>
                </div>
            </CodyAlert>
        )
    }

    return null
}
