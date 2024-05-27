import React, { useEffect } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { PageHeader, Text, H3, useSearchParameters } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { CodyAlert } from '../components/CodyAlert'
import { WhiteIcon } from '../components/WhiteIcon'
import { requestSSC } from '../util'

interface CodyAcceptInvitePageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedCodyAcceptInvitePage: React.FunctionComponent<CodyAcceptInvitePageProps> = ({
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.invites.accept', 'view')
    }, [telemetryRecorder])

    const navigate = useNavigate()

    // Process query params
    const parameters = useSearchParameters()
    const teamId = parameters.get('teamID')
    const inviteId = parameters.get('inviteID')
    const [loading, setLoading] = React.useState(true)
    const [errorMessage, setErrorMessage] = React.useState<string | null>(null)

    useEffect(() => {
        async function postAcceptInvite(): Promise<void> {
            if (!teamId || !inviteId) {
                setErrorMessage('Invalid invite ID or team ID')
                setLoading(false)
                return
            }
            const response = await requestSSC(`/team/${teamId}/invites/${inviteId}/accept`, 'POST')
            setLoading(false)
            if (response.ok) {
                // Wait a second before navigating to the manage team page so that the user sees the success alert.
                await new Promise(resolve => setTimeout(resolve, 1000))
                navigate('/cody/team/manage?welcome=1')
            } else {
                setErrorMessage(await response.text())
            }
        }

        void postAcceptInvite()
    }, [inviteId, navigate, teamId])

    return (
        <Page className={classNames('d-flex flex-column')}>
            <PageTitle title="Manage Cody team" />
            <PageHeader className="mb-4 mt-4">
                <PageHeader.Heading as="h2" styleAs="h1">
                    <div className="d-inline-flex align-items-center">
                        <WhiteIcon name="mdi-account-multiple-plus-gradient" className="mr-3" />
                        Join Cody Pro team
                    </div>
                </PageHeader.Heading>
            </PageHeader>

            {errorMessage ? (
                <CodyAlert variant="error">
                    <H3>We couldn't accept the invite.</H3>
                    <Text size="small" className="text-muted mb-0">
                        {errorMessage}
                    </Text>
                </CodyAlert>
            ) : null}

            {!loading && !errorMessage ? (
                <CodyAlert variant="greenSuccess">
                    <H3>Invite accepted!</H3>
                    <Text size="small" className="mb-0">
                        You are now a member of the team. We'll now redirect you to the team page.
                    </Text>
                </CodyAlert>
            ) : null}
        </Page>
    )
}

export const CodyAcceptInvitePage = withAuthenticatedUser(AuthenticatedCodyAcceptInvitePage)
