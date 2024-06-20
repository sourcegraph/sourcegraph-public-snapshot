import React, { useCallback, useEffect } from 'react'

import { mdiCreditCardOutline, mdiPlusThick } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { ButtonLink, H1, H2, Icon, Link, PageHeader, Text, useSearchParameters, Button } from '@sourcegraph/wildcard'

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
import { CodyProRoutes } from '../codyProRoutes'
import { CodyAlert } from '../components/CodyAlert'
import { ProIcon } from '../components/CodyIcon'
import { PageHeaderIcon } from '../components/PageHeaderIcon'
import { AcceptInviteBanner } from '../invites/AcceptInviteBanner'
import { InviteUsers } from '../invites/InviteUsers'
import { isCodyEnabled } from '../isCodyEnabled'
import { CodyOnboarding, type IEditor } from '../onboarding/CodyOnboarding'
import { USER_CODY_PLAN, USER_CODY_USAGE } from '../subscription/queries'
import { getManageSubscriptionPageURL } from '../util'

import { useSubscriptionSummary } from './api/react-query/subscriptions'
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

    // The cody_client_user URL query param is added by the VS Code & JetBrains
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

    const { data, error: dataError, refetch } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

    const { data: usageData, error: usageDateError } = useQuery<UserCodyUsageResult, UserCodyUsageVariables>(
        USER_CODY_USAGE,
        {}
    )

    const subscriptionSummaryQueryResult = useSubscriptionSummary()
    const isAdmin = subscriptionSummaryQueryResult?.data?.userRole === 'admin'

    const [selectedEditor, setSelectedEditor] = React.useState<IEditor | null>(null)
    const [selectedEditorStep, setSelectedEditorStep] = React.useState<EditorStep | null>(null)

    const subscription = data?.currentUser?.codySubscription

    useEffect(() => {
        if (!!data && !data?.currentUser) {
            navigate(`/sign-in?returnTo=${CodyProRoutes.Manage}`)
        }
    }, [data, navigate])

    const getTeamInviteButton = (): JSX.Element | null => {
        const isSoloUser = subscriptionSummaryQueryResult?.data?.teamMaxMembers === 1
        const hasFreeSeats = subscriptionSummaryQueryResult?.data
            ? subscriptionSummaryQueryResult.data.teamMaxMembers >
              subscriptionSummaryQueryResult.data.teamCurrentMembers
            : false
        const targetUrl = hasFreeSeats ? CodyProRoutes.ManageTeam : `${CodyProRoutes.NewProSubscription}?addSeats=1`
        const label = isSoloUser || hasFreeSeats ? 'Invite co-workers' : 'Add seats'

        if (!subscriptionSummaryQueryResult?.data) {
            return null
        }

        return (
            <Button as={Link} to={targetUrl} variant="success" className="text-nowrap">
                <Icon aria-hidden={true} svgPath={mdiPlusThick} /> {label}
            </Button>
        )
    }

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
                <AcceptInviteBanner onSuccess={refetch} />
                {welcomeToPro && (
                    <CodyAlert variant="greenCodyPro">
                        <H2 className="mt-4">Welcome to Cody Pro</H2>
                        <Text size="small" className="mb-0">
                            You now have Cody Pro with access to unlimited autocomplete, chats, and commands.
                        </Text>
                    </CodyAlert>
                )}
                <PageHeader
                    className="my-4 d-inline-flex align-items-center"
                    actions={isAdmin && <div className="d-flex">{getTeamInviteButton()}</div>}
                >
                    <PageHeader.Heading as="h1" className="text-3xl font-medium">
                        <PageHeaderIcon name="dashboard" className="mr-3" />
                        <Text as="span">Dashboard</Text>
                    </PageHeader.Heading>
                </PageHeader>

                {isAdmin && !!subscriptionSummaryQueryResult.data && (
                    <InviteUsers
                        telemetryRecorder={telemetryRecorder}
                        subscriptionSummary={subscriptionSummaryQueryResult.data}
                    />
                )}

                {!isUserOnProTier && <UpgradeToProBanner onClick={onClickUpgradeToProCTA} />}

                <div className={classNames('p-4 border bg-1 mt-3', styles.container)}>
                    <div className="d-flex justify-content-between align-items-center border-bottom pb-3">
                        <div>
                            <H2>My subscription</H2>
                            <Text className="text-muted mb-0">
                                {isUserOnProTier ? (
                                    'You are on the Pro tier.'
                                ) : (
                                    <span>
                                        You are on the Free tier.{' '}
                                        <Link to={CodyProRoutes.Subscription}>Upgrade to the Pro tier.</Link>
                                    </span>
                                )}
                            </Text>
                        </div>
                        {isAdmin && (
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
    <CodyAlert variant="purpleCodyPro">
        <div className="d-flex justify-content-between align-items-center p-4">
            <div>
                <H1>
                    Get unlimited help with <span className={styles.codyProGradientText}>Cody Pro</span>
                </H1>
                <ul className="pl-4 mb-0">
                    <li>Unlimited autocompletions</li>
                    <li>Unlimited chat messages</li>
                </ul>
            </div>
            <div>
                <ButtonLink to={CodyProRoutes.Subscription} variant="primary" size="sm" onClick={onClick}>
                    <ProIcon className="mr-1" />
                    Upgrade now
                </ButtonLink>
            </div>
        </div>
    </CodyAlert>
)
