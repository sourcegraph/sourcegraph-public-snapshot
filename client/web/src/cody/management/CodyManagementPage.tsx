import React, { useCallback, useEffect } from 'react'

import { mdiCreditCardOutline } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { ButtonLink, H1, H2, Icon, Link, PageHeader, Text, useSearchParameters } from '@sourcegraph/wildcard'

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
import { AcceptInviteBanner } from '../invites/AcceptInviteBanner'
import { isCodyEnabled } from '../isCodyEnabled'
import { CodyOnboarding, type IEditor } from '../onboarding/CodyOnboarding'
import { USER_CODY_PLAN, USER_CODY_USAGE } from '../subscription/queries'
import { getManageSubscriptionPageURL } from '../util'

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

    const { data, error: dataError, refetch } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

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
                <CodyAlert variant="purpleCodyPro">
                    <H1 as="p" className="mb-2">
                        Issue with invite
                    </H1>
                    <Text className="mb-0">
                        You've been invited to a Cody Pro team by john.doe@email.com.
                        <br />
                        You cannot accept this invite as as you are already on this team.
                    </Text>
                </CodyAlert>
                <AcceptInviteBanner onSuccess={refetch} />
                {welcomeToPro && (
                    <CodyAlert variant="greenCodyPro">
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

const IssueIcon = () => (
    <svg
        width="124"
        height="148"
        viewBox="0 0 124 148"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style={{ position: 'absolute' }}
    >
        <g filter="url(#filter0_d_5472_15314)">
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M11 4.5C7.68555 4.5 5 7.7189 5 11.6897V136.31C5 140.281 7.68555 143.5 11 143.5H113C116.314 143.5 119 140.281 119 136.31V11.6897C119 7.7189 116.314 4.5 113 4.5H11ZM55.8203 15.673C54.168 15.673 52.8301 17.2773 52.8301 19.2563C52.8301 21.2353 54.168 22.8397 55.8203 22.8397H68.1816C69.832 22.8397 71.1719 21.2353 71.1719 19.2563C71.1719 17.2773 69.832 15.673 68.1816 15.673H55.8203Z"
                fill="#D9D9D9"
            />
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M11 4.5C7.68555 4.5 5 7.7189 5 11.6897V136.31C5 140.281 7.68555 143.5 11 143.5H113C116.314 143.5 119 140.281 119 136.31V11.6897C119 7.7189 116.314 4.5 113 4.5H11ZM55.8203 15.673C54.168 15.673 52.8301 17.2773 52.8301 19.2563C52.8301 21.2353 54.168 22.8397 55.8203 22.8397H68.1816C69.832 22.8397 71.1719 21.2353 71.1719 19.2563C71.1719 17.2773 69.832 15.673 68.1816 15.673H55.8203Z"
                fill="#EFF2F5"
            />
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M11 4.5C7.68555 4.5 5 7.7189 5 11.6897V136.31C5 140.281 7.68555 143.5 11 143.5H113C116.314 143.5 119 140.281 119 136.31V11.6897C119 7.7189 116.314 4.5 113 4.5H11ZM55.8203 15.673C54.168 15.673 52.8301 17.2773 52.8301 19.2563C52.8301 21.2353 54.168 22.8397 55.8203 22.8397H68.1816C69.832 22.8397 71.1719 21.2353 71.1719 19.2563C71.1719 17.2773 69.832 15.673 68.1816 15.673H55.8203Z"
                fill="white"
                fillOpacity="0.2"
            />
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M11 4.5C7.68555 4.5 5 7.7189 5 11.6897V136.31C5 140.281 7.68555 143.5 11 143.5H113C116.314 143.5 119 140.281 119 136.31V11.6897C119 7.7189 116.314 4.5 113 4.5H11ZM55.8203 15.673C54.168 15.673 52.8301 17.2773 52.8301 19.2563C52.8301 21.2353 54.168 22.8397 55.8203 22.8397H68.1816C69.832 22.8397 71.1719 21.2353 71.1719 19.2563C71.1719 17.2773 69.832 15.673 68.1816 15.673H55.8203Z"
                fill="url(#paint0_linear_5472_15314)"
            />
            <path
                d="M5.2 11.6897C5.2 7.79394 7.82837 4.7 11 4.7H113C116.172 4.7 118.8 7.79394 118.8 11.6897V136.31C118.8 140.206 116.172 143.3 113 143.3H11C7.82837 143.3 5.2 140.206 5.2 136.31V11.6897ZM55.8203 15.473C54.0251 15.473 52.6301 17.2023 52.6301 19.2563C52.6301 21.3103 54.0251 23.0397 55.8203 23.0397H68.1816C69.9748 23.0397 71.3719 21.3104 71.3719 19.2563C71.3719 17.2022 69.9748 15.473 68.1816 15.473H55.8203Z"
                stroke="black"
                strokeOpacity="0.16"
                strokeWidth="0.4"
            />
            <path
                d="M63.667 68.834H60.3337V62.1673H63.667V68.834ZM63.667 75.5007H60.3337V72.1673H63.667V75.5007ZM43.667 80.5007H80.3337L62.0003 48.834L43.667 80.5007Z"
                fill="#DBE2F0"
            />
        </g>
        <defs>
            <filter
                id="filter0_d_5472_15314"
                x="0.5"
                y="0"
                width="123"
                height="148"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feMorphology radius="1" operator="dilate" in="SourceAlpha" result="effect1_dropShadow_5472_15314" />
                <feOffset />
                <feGaussianBlur stdDeviation="1.75" />
                <feColorMatrix type="matrix" values="0 0 0 0 0.141522 0 0 0 0 0.159783 0 0 0 0 0.21 0 0 0 0.31 0" />
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow_5472_15314" />
                <feBlend mode="normal" in="SourceGraphic" in2="effect1_dropShadow_5472_15314" result="shape" />
            </filter>
            <linearGradient
                id="paint0_linear_5472_15314"
                x1="62.2346"
                y1="136.573"
                x2="62.2346"
                y2="4.5"
                gradientUnits="userSpaceOnUse"
            >
                <stop offset="0.433329" stopColor="white" />
                <stop offset="1" stopColor="white" stopOpacity="0" />
            </linearGradient>
        </defs>
    </svg>
)
