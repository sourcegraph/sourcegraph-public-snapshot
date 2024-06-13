import React, { useCallback, useEffect, useState } from 'react'

import { mdiCreditCardOutline } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Button,
    ButtonLink,
    H1,
    H2,
    H3,
    Icon,
    Link,
    PageHeader,
    Text,
    useSearchParameters,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import {
    type UserCodyPlanResult,
    type UserCodyPlanVariables,
    type UserCodyUsageResult,
    type UserCodyUsageVariables,
    CodySubscriptionPlan,
} from '../../graphql-operations'
import { CodyAlert } from '../components/CodyAlert'
import { CodyProIcon, DashboardIcon } from '../components/CodyIcon'
import { useInviteParams, UserInviteStatus, useUserInviteStatus } from '../invites/AcceptInvitePage'
import { isCodyEnabled } from '../isCodyEnabled'
import { CodyOnboarding, type IEditor } from '../onboarding/CodyOnboarding'
import { USER_CODY_PLAN, USER_CODY_USAGE } from '../subscription/queries'
import { getManageSubscriptionPageURL } from '../util'

import { useAcceptInvite, useCancelInvite, useInvite } from './api/react-query/invites'
import { SubscriptionStats } from './SubscriptionStats'
import { UseCodyInEditorSection } from './UseCodyInEditorSection'

import styles from './CodyManagementPage.module.scss'

interface CodyManagementPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
}

export enum EditorStep {
    SetupInstructions = 0,
    CodyFeatures = 1,
}

export const CodyManagementPage: React.FunctionComponent<CodyManagementPageProps> = ({
    authenticatedUser,
    telemetryRecorder,
}) => {
    const navigate = useNavigate()
    const parameters = useSearchParameters()

    const utm_source = parameters.get('utm_source')
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.management', 'view')
    }, [utm_source, telemetryRecorder])

    // The cody_client_user URL query param is added by the VS Code & Jetbrains
    // extensions. We redirect them to a "switch account" screen if they are
    // logged into their IDE as a different user account than their browser.
    const codyClientUser = parameters.get('cody_client_user')
    const accountSwitchRequired = !!codyClientUser && authenticatedUser && authenticatedUser.username !== codyClientUser
    useEffect(() => {
        if (accountSwitchRequired) {
            navigate(`/cody/switch-account/${codyClientUser}`)
        }
    }, [accountSwitchRequired, codyClientUser, navigate])

    const welcomeToPro = parameters.get('welcome') === '1'

    const { data, error: dataError } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

    const { data: usageData, error: usageDateError } = useQuery<UserCodyUsageResult, UserCodyUsageVariables>(
        USER_CODY_USAGE,
        {}
    )

    const [selectedEditor, setSelectedEditor] = React.useState<IEditor | null>(null)
    const [selectedEditorStep, setSelectedEditorStep] = React.useState<EditorStep | null>(null)

    const subscription = data?.currentUser?.codySubscription

    useEffect(() => {
        if (!!data && !data?.currentUser) {
            navigate('/sign-in?returnTo=/cody/manage')
        }
    }, [data, navigate])

    const onClickUpgradeToProCTA = useCallback(() => {
        telemetryRecorder.recordEvent('cody.management.upgradeToProCTA', 'click')
    }, [telemetryRecorder])

    if (accountSwitchRequired) {
        return null
    }

    if (dataError || usageDateError) {
        throw dataError || usageDateError
    }

    if (!isCodyEnabled() || !subscription) {
        return null
    }

    const isUserOnProTier = subscription.plan === CodySubscriptionPlan.PRO

    return (
        <>
            <Page className={classNames('d-flex flex-column')}>
                <PageTitle title="Dashboard" />
                <AcceptInviteBanner />
                {welcomeToPro && (
                    <CodyAlert variant="purpleCodyPro">
                        <H2 className="mt-4">Welcome to Cody Pro</H2>
                        <Text size="small" className="mb-0">
                            You now have Cody Pro with access to unlimited autocomplete, chats, and commands.
                        </Text>
                    </CodyAlert>
                )}
                <PageHeader className="mb-4 mt-4">
                    <PageHeader.Heading as="h2" styleAs="h1">
                        <div className="d-inline-flex align-items-center">
                            <DashboardIcon className="mr-2" /> Dashboard
                        </div>
                    </PageHeader.Heading>
                </PageHeader>

                {!isUserOnProTier && <UpgradeToProBanner onClick={onClickUpgradeToProCTA} />}

                <div className={classNames('p-4 border bg-1 mt-4', styles.container)}>
                    <div className="d-flex justify-content-between align-items-center border-bottom pb-3">
                        <div>
                            <H2>My subscription</H2>
                            <Text className="text-muted mb-0">
                                {isUserOnProTier ? (
                                    'You are on the Pro tier.'
                                ) : (
                                    <span>
                                        You are on the Free tier.{' '}
                                        <Link to="/cody/subscription">Upgrade to the Pro tier.</Link>
                                    </span>
                                )}
                            </Text>
                        </div>
                        {isUserOnProTier && (
                            <div>
                                <ButtonLink
                                    variant="primary"
                                    size="sm"
                                    to={getManageSubscriptionPageURL()}
                                    onClick={() => {
                                        telemetryRecorder.recordEvent('cody.manageSubscription', 'click')
                                    }}
                                >
                                    <Icon svgPath={mdiCreditCardOutline} className="mr-1" aria-hidden={true} />
                                    Manage subscription
                                </ButtonLink>
                            </div>
                        )}
                    </div>
                    <SubscriptionStats {...{ subscription, usageData }} />
                </div>

                <UseCodyInEditorSection
                    {...{
                        selectedEditor,
                        setSelectedEditor,
                        selectedEditorStep,
                        setSelectedEditorStep,
                        isUserOnProTier,
                        telemetryRecorder,
                    }}
                />
            </Page>
            <CodyOnboarding authenticatedUser={authenticatedUser} telemetryRecorder={telemetryRecorder} />
        </>
    )
}

const UpgradeToProBanner: React.FunctionComponent<{
    onClick: () => void
}> = ({ onClick }) => (
    <div className={classNames('d-flex justify-content-between align-items-center p-4', styles.upgradeToProBanner)}>
        <div>
            <H1>
                Become limitless with
                <CodyProIcon className="ml-1" />
            </H1>
            <ul className="pl-4 mb-0">
                <li>Unlimited autocompletions</li>
                <li>Unlimited chat messages</li>
            </ul>
        </div>
        <div>
            <ButtonLink to="/cody/subscription" variant="primary" size="sm" onClick={onClick}>
                Upgrade
            </ButtonLink>
        </div>
    </div>
)

const AcceptInviteBanner: React.FC = () => {
    const { params: inviteParams, clear: clearInviteParams } = useInviteParams()
    const userInviteStatus = useUserInviteStatus()

    const inviteQuery = useInvite(inviteParams)
    const acceptInviteMutation = useAcceptInvite()
    const cancelInviteMutation = useCancelInvite()

    // Keep track only of initial non-undefined userInviteStatus.
    // userInviteStatus depends on search params and backend state, so it can change over time
    // (e.g., accept invite and instead of UserInviteStatus.NoCurrentTeam user has UserInviteStatus.InvitedTeamMember status).
    // TODO: this implemetation is far from ideal. Refactor it in the current PR.
    const [status, setStatus] = useState<UserInviteStatus>()
    useEffect(() => {
        setStatus(s => (s === undefined ? userInviteStatus : s))
    }, [userInviteStatus])

    if (inviteParams === undefined && acceptInviteMutation.isIdle && cancelInviteMutation.isIdle) {
        return null
    }

    // TODO: handle invite states other than "sent"

    switch (status) {
        case UserInviteStatus.NoCurrentTeam:
        case UserInviteStatus.AnotherTeamMember: {
            const sentBy = inviteQuery.data?.sentBy

            switch (acceptInviteMutation.status) {
                case 'error': {
                    return (
                        <CodyAlert variant="purple">
                            <H3 className="mt-4">Failed to accept invite</H3>
                            <Text className="text-danger">{acceptInviteMutation.error.message}</Text>
                        </CodyAlert>
                    )
                }
                case 'success': {
                    return (
                        <CodyAlert variant="purple">
                            <H3 className="mt-4">Success</H3>
                            <Text>You joined the new Cody Pro team.</Text>
                        </CodyAlert>
                    )
                }
                default: {
                    return (
                        <CodyAlert variant="purple">
                            <H3 className="mt-4">Join new Cody Pro team?</H3>
                            <Text>You've been invited to a new Cody Pro team{sentBy ? ` by ${sentBy}` : ''}.</Text>
                            <Text>
                                {userInviteStatus === UserInviteStatus.NoCurrentTeam
                                    ? 'You will get unlimited autocompletions chat messages.'
                                    : 'This will terminate your current Cody Pro plan, and place you on the new Cody Pro team. You will not lose access to your Cody Pro benefits.'}
                            </Text>
                            <div>
                                <Button
                                    onClick={() =>
                                        acceptInviteMutation.mutate(inviteParams!, { onSuccess: clearInviteParams })
                                    }
                                >
                                    Accept
                                </Button>
                                <Button
                                    variant="secondary"
                                    onClick={() =>
                                        cancelInviteMutation.mutate(inviteParams!, { onSettled: clearInviteParams })
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
            if (inviteParams && cancelInviteMutation.isIdle) {
                void cancelInviteMutation.mutate(inviteParams, { onSettled: clearInviteParams })
            }
            return (
                <CodyAlert variant="purple">
                    <H3 className="mt-4">You are already memeber of the team.</H3>
                    <Text>You've been invited to a new Cody Pro team by rob@acmecorp.com.</Text>
                    <Text>This invite will be canceled.</Text>
                </CodyAlert>
            )
        }
        case UserInviteStatus.Error: {
            if (inviteParams && cancelInviteMutation.isIdle) {
                void cancelInviteMutation.mutate(inviteParams, { onSettled: clearInviteParams })
            }
            return (
                <CodyAlert variant="error">
                    <H3 className="mt-4">Failed to process an invite.</H3>
                    <Text>Invite link is broken or failed to fetch subscription data.</Text>
                    <Text>
                        Ask an admin for another invite. If the problem persist, reach out to support@sourcegraph.com.
                    </Text>
                    <Text>This invite will be canceled.</Text>
                </CodyAlert>
            )
        }
        default: {
            return null
        }
    }
}
