import React, { useCallback, useEffect } from 'react'

import { mdiCreditCardOutline, mdiHelpCircleOutline, mdiPlusThick } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Button, ButtonLink, H2, H3, Icon, Link, PageHeader, Text, useSearchParameters } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import {
    CodySubscriptionPlan,
    type UserCodyPlanResult,
    type UserCodyPlanVariables,
    type UserCodyUsageResult,
    type UserCodyUsageVariables,
} from '../../graphql-operations'
import { CodyProRoutes } from '../codyProRoutes'
import { CodyAlert } from '../components/CodyAlert'
import { PageHeaderIcon } from '../components/PageHeaderIcon'
import { AcceptInviteBanner } from '../invites/AcceptInviteBanner'
import { InviteUsers } from '../invites/InviteUsers'
import { USER_CODY_PLAN, USER_CODY_USAGE } from '../subscription/queries'
import { getManageSubscriptionPageURL, isEmbeddedCodyProUIEnabled } from '../util'

import { useSubscriptionSummary } from './api/react-query/subscriptions'
import { SubscriptionStats } from './SubscriptionStats'
import { CodyEditorsAndClients } from './UseCodyInEditorSection'

import styles from './CodyManagementPage.module.scss'

interface CodyManagementPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
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

    const subscription = data?.currentUser?.codySubscription

    useEffect(() => {
        if (!!data && !data?.currentUser) {
            navigate(`/sign-in?returnTo=${CodyProRoutes.Manage}`)
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

    if (!window.context?.codyEnabledForCurrentUser || !subscription) {
        return null
    }

    // Generate invite-related widgets for the page.
    // The returned widgets depend on the user's role and subscription status.
    // Only applicable when embedded Cody Pro UI is enabled.
    const inviteWidgets = ((): {
        banner?: React.ReactElement
        link?: React.ReactElement
        form?: React.ReactElement
    } => {
        // Invites flow is supported only for embedded Cody UI.
        if (!isEmbeddedCodyProUIEnabled()) {
            return {}
        }

        const banner = <AcceptInviteBanner onSuccess={refetch} />

        const subscriptionSummary = subscriptionSummaryQueryResult.data
        // User is not admin: only banner widget is available from the invite flow (user with any role may get an invite).
        if (!subscriptionSummary || subscriptionSummary.userRole !== 'admin') {
            return { banner }
        }

        // User is admin: define link and form widgets.

        const form = <InviteUsers telemetryRecorder={telemetryRecorder} subscriptionSummary={subscriptionSummary} />

        const isSoloUser = subscriptionSummary.teamMaxMembers === 1
        const hasFreeSeats = subscriptionSummary.teamMaxMembers > subscriptionSummary.teamCurrentMembers
        const targetUrl = hasFreeSeats ? CodyProRoutes.ManageTeam : `${CodyProRoutes.NewProSubscription}?addSeats=1`
        const label = isSoloUser || hasFreeSeats ? 'Invite co-workers' : 'Add seats'
        const link = (
            <Button as={Link} to={targetUrl} variant="success" className="text-nowrap">
                <Icon aria-hidden={true} svgPath={mdiPlusThick} /> {label}
            </Button>
        )

        return { banner, link, form }
    })()

    const pageHeaderLink: React.ReactNode = (() => {
        // If invites link widget is defined, render it.
        if (inviteWidgets.link) {
            return inviteWidgets.link
        }

        // User already has a Pro subscription - render a link to the manage subscription page.
        if (subscription.plan === CodySubscriptionPlan.PRO) {
            return (
                <ButtonLink
                    variant="primary"
                    to={getManageSubscriptionPageURL()}
                    onClick={() => {
                        telemetryRecorder.recordEvent('cody.manageSubscription', 'click')
                    }}
                >
                    <Icon svgPath={mdiCreditCardOutline} className="mr-1" aria-hidden={true} />
                    Manage subscription
                </ButtonLink>
            )
        }

        // User is on a free plan. Render a link to the subscriptions page.
        return (
            <ButtonLink to={CodyProRoutes.Subscription} variant="primary" onClick={onClickUpgradeToProCTA}>
                Upgrade plan
            </ButtonLink>
        )
    })()

    return (
        <Page className={classNames('d-flex flex-column')}>
            <PageTitle title="Dashboard" />

            {inviteWidgets.banner}

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
                actions={
                    <div className="d-flex flex-column flex-gap-2">
                        {pageHeaderLink}
                        <Link
                            to="https://help.sourcegraph.com"
                            target="_blank"
                            rel="noreferrer"
                            className="text-muted text-center text-sm"
                        >
                            <Icon svgPath={mdiHelpCircleOutline} className="mr-1" aria-hidden={true} />
                            Help &amp; community
                        </Link>
                    </div>
                }
            >
                <PageHeader.Heading as="h1" className="text-3xl font-medium">
                    <PageHeaderIcon name="dashboard" className="mr-3" />
                    <Text as="span">Cody dashboard</Text>
                </PageHeader.Heading>
            </PageHeader>

            {inviteWidgets.form}

            <div className={classNames('border bg-1 mb-2', styles.container)}>
                <SubscriptionStats
                    subscription={subscription}
                    usageData={usageData}
                    onClickUpgradeToProCTA={onClickUpgradeToProCTA}
                />
            </div>

            <H3 className="mt-3 text-muted">Use Cody...</H3>
            <div className={classNames('border bg-1 mb-2', styles.container)}>
                <CodyEditorsAndClients telemetryRecorder={telemetryRecorder} />
            </div>

            <div className="pb-3" />
        </Page>
    )
}
