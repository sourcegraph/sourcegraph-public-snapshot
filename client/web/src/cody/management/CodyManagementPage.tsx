import React, { useEffect } from 'react'

import { mdiHelpCircleOutline, mdiInformationOutline, mdiOpenInNew, mdiCreditCardOutline } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import {
    ButtonLink,
    H1,
    H2,
    H3,
    H4,
    H5,
    Icon,
    Link,
    LoadingSpinner,
    Modal,
    PageHeader,
    Text,
    useSearchParameters,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import type {
    ChangeCodyPlanResult,
    ChangeCodyPlanVariables,
    UserCodyPlanResult,
    UserCodyPlanVariables,
    UserCodyUsageResult,
    UserCodyUsageVariables,
} from '../../graphql-operations'
import { CodySubscriptionStatus, CodySubscriptionPlan } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { EventName } from '../../util/constants'
import {
    CodyProIcon,
    AutocompletesIcon,
    ChatMessagesIcon,
    TrialPeriodIcon,
    DashboardIcon,
} from '../components/CodyIcon'
import { useArePaymentsEnabled, useHasTrialEnded } from '../featureFlags'
import { isCodyEnabled } from '../isCodyEnabled'
import { CodyOnboarding, editorGroups, type IEditor } from '../onboarding/CodyOnboarding'
import { ProTierIcon, useCodyPaymentsUrl } from '../subscription/CodySubscriptionPage'
import { CHANGE_CODY_PLAN, USER_CODY_PLAN, USER_CODY_USAGE } from '../subscription/queries'

import styles from './CodyManagementPage.module.scss'

interface CodyManagementPageProps {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const CodyManagementPage: React.FunctionComponent<CodyManagementPageProps> = ({
    isSourcegraphDotCom,
    authenticatedUser,
}) => {
    const parameters = useSearchParameters()

    const utm_source = parameters.get('utm_source')

    const arePaymentsEnabled = useArePaymentsEnabled()
    const hasTrialEnded = useHasTrialEnded()

    useEffect(() => {
        eventLogger.log(EventName.CODY_MANAGEMENT_PAGE_VIEWED, { utm_source })
    }, [utm_source])

    const { data, error: dataError } = useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

    const { data: usageData, error: usageDateError } = useQuery<UserCodyUsageResult, UserCodyUsageVariables>(
        USER_CODY_USAGE,
        {}
    )

    const stats = usageData?.currentUser
    const codyCurrentPeriodChatLimit = stats?.codyCurrentPeriodChatLimit || 0
    const codyCurrentPeriodChatUsage = stats?.codyCurrentPeriodChatUsage || 0
    const codyCurrentPeriodCodeLimit = stats?.codyCurrentPeriodCodeLimit || 0
    const codyCurrentPeriodCodeUsage = stats?.codyCurrentPeriodCodeUsage || 0

    const [changeCodyPlan] = useMutation<ChangeCodyPlanResult, ChangeCodyPlanVariables>(CHANGE_CODY_PLAN)

    const [selectedEditor, setSelectedEditor] = React.useState<IEditor | null>(null)
    const [selectedEditorStep, setSelectedEditorStep] = React.useState<number | null>(null)

    const enrollPro = parameters.get('pro') === 'true'

    const subscription = data?.currentUser?.codySubscription

    const codyPaymentsUrl = useCodyPaymentsUrl()
    const manageSubscriptionRedirectURL = `${codyPaymentsUrl}/cody/subscription`

    useEffect(() => {
        if (!arePaymentsEnabled && enrollPro && data?.currentUser && subscription?.plan !== CodySubscriptionPlan.PRO) {
            void changeCodyPlan({ variables: { pro: true, id: data?.currentUser?.id } })
        }
    }, [arePaymentsEnabled, data?.currentUser, changeCodyPlan, enrollPro, subscription])

    const navigate = useNavigate()

    useEffect(() => {
        if (!!data && !data?.currentUser) {
            navigate('/sign-in?returnTo=/cody/manage')
        }
    }, [data, navigate])

    if (dataError || usageDateError) {
        throw dataError || usageDateError
    }

    if (!isCodyEnabled() || !isSourcegraphDotCom || !subscription) {
        return null
    }

    const codeLimitReached = codyCurrentPeriodCodeUsage >= codyCurrentPeriodCodeLimit && codyCurrentPeriodCodeLimit > 0
    const chatLimitReached = codyCurrentPeriodChatUsage >= codyCurrentPeriodChatLimit && codyCurrentPeriodChatLimit > 0
    const userIsOnProTier =
        subscription.plan === CodySubscriptionPlan.PRO &&
        !(subscription.status === CodySubscriptionStatus.TRIALING && subscription.cancelAtPeriodEnd)
    const hasNotAddedCreditCard = subscription.status === CodySubscriptionStatus.PENDING

    // Flag usage limits as resetting based on the current subscription's billing cycle.
    //
    // BUG: The usage limit refresh should be independent of a user's subscription data.
    //      e.g. if we offered an annual billing plan, we'd want to reset usage more often.
    //      sourcegraph#59990 is related, and required for the times to line up with the
    //      behavior from Cody Gateway.
    //
    // BUG: If the subscription is canceled, this will be in the past and therefore invalid.
    //      This data should be fetched from the SSC backend, and like above, separeate
    //      from the user's subscription billing cycle.
    const usageRefreshTime = subscription.currentPeriodEndAt

    // Time when the user's current subscription will end.
    //
    // BUG: If the subscription is in the canceled state, this will be in the past. We need
    //      to update the UI to simply say "subscription canceled" or "you are on the free"
    //      plan, you don't have any subscription billing cycle anchors".
    //
    let codyProSubscriptionEndTime = subscription.currentPeriodEndAt

    // Correct the situation where the user is on a Cody Pro free trial, but hasn't entered
    // any subscription information into the SSC frontend. This would mean that their free
    // trial is coming to an end on ~2/21. We need the UI to reflect this, however, because
    // we are overloading `currentPeriodEnd` for usageRefreshTime, we do not return the
    // correct value from the backend. So we separate it out into a separate variable and
    // change its value accordingly.
    const freeTrialEndString = 'Until Feb 21, 2024'
    if (!hasTrialEnded && userIsOnProTier) {
        codyProSubscriptionEndTime = new Date(2024, 2, 21, 12, 0, 0).toISOString()
    }

    return (
        <>
            <Page className={classNames('d-flex flex-column')}>
                <PageTitle title="Dashboard" />
                <PageHeader className="mb-4 mt-4">
                    <PageHeader.Heading as="h2" styleAs="h1">
                        <div className="d-inline-flex align-items-center">
                            <DashboardIcon className="mr-2" /> Dashboard
                        </div>
                    </PageHeader.Heading>
                </PageHeader>

                <UpgradeToProBanner userIsOnProTier={userIsOnProTier} />
                <DoNotLoseCodyProBanner
                    userIsOnProTier={userIsOnProTier}
                    arePaymentsEnabled={arePaymentsEnabled}
                    hasTrialEnded={hasTrialEnded}
                    subscriptionStatus={subscription.status}
                />

                <div className={classNames('p-4 border bg-1 mt-4', styles.container)}>
                    <div className="d-flex justify-content-between align-items-center border-bottom pb-3">
                        <div>
                            <H2>My subscription</H2>
                            <Text className="text-muted mb-0">
                                {userIsOnProTier ? (
                                    hasTrialEnded && hasNotAddedCreditCard ? (
                                        'Your Cody Pro trial has ended.'
                                    ) : (
                                        'You are on the Pro tier.'
                                    )
                                ) : (
                                    <span>
                                        You are on the Free tier.{' '}
                                        <Link to="/cody/subscription">Upgrade to the Pro tier.</Link>
                                    </span>
                                )}
                            </Text>
                        </div>
                        {userIsOnProTier && (
                            <div>
                                <ButtonLink
                                    variant="primary"
                                    size="sm"
                                    onClick={() => {
                                        eventLogger.log(EventName.CODY_MANAGE_SUBSCRIPTION_CLICKED)
                                        window.location.href = manageSubscriptionRedirectURL
                                    }}
                                >
                                    <Icon svgPath={mdiCreditCardOutline} className="mr-1" aria-hidden={true} />
                                    Manage subscription
                                </ButtonLink>
                            </div>
                        )}
                    </div>
                    <div className={classNames('d-flex align-items-center mt-3', styles.responsiveContainer)}>
                        <div className="d-flex flex-column align-items-center flex-grow-1 p-3">
                            {userIsOnProTier && !(hasTrialEnded && hasNotAddedCreditCard) ? (
                                <ProTierIcon />
                            ) : (
                                <Text className={classNames(styles.planName, 'mb-0')}>Free</Text>
                            )}
                            <Text className="text-muted mb-0" size="small">
                                tier
                            </Text>
                            {userIsOnProTier && subscription.cancelAtPeriodEnd && (
                                <Text className="text-muted mb-0 mt-4" size="small">
                                    Subscription ends <Timestamp date={codyProSubscriptionEndTime} />
                                </Text>
                            )}
                        </div>
                        <div className="d-flex flex-column align-items-center flex-grow-1 p-3 border-left border-right">
                            <AutocompletesIcon />
                            <div className="mb-2 mt-3">
                                {subscription.applyProRateLimits ? (
                                    <Text weight="bold" className={classNames('d-inline mb-0', styles.counter)}>
                                        Unlimited
                                    </Text>
                                ) : usageData?.currentUser ? (
                                    <>
                                        <Text
                                            weight="bold"
                                            className={classNames(
                                                'd-inline mb-0',
                                                styles.counter,
                                                codeLimitReached ? 'text-danger' : 'text-muted'
                                            )}
                                        >
                                            {Math.min(codyCurrentPeriodCodeUsage, codyCurrentPeriodCodeLimit)} /
                                        </Text>{' '}
                                        <Text
                                            className={classNames(
                                                'd-inline b-0',
                                                codeLimitReached ? 'text-danger' : 'text-muted'
                                            )}
                                            size="small"
                                        >
                                            {codyCurrentPeriodCodeLimit}
                                        </Text>
                                    </>
                                ) : (
                                    <LoadingSpinner />
                                )}
                            </div>
                            <H4 className={classNames('mb-0', codeLimitReached ? 'text-danger' : 'text-muted')}>
                                Autocomplete suggestions
                            </H4>
                            {!subscription.applyProRateLimits &&
                                (codeLimitReached ? (
                                    <Text className="text-danger mb-0" size="small">
                                        Renews in <Timestamp date={usageRefreshTime} />
                                    </Text>
                                ) : (
                                    <Text className="text-muted mb-0" size="small">
                                        this month
                                    </Text>
                                ))}
                        </div>
                        <div className="d-flex flex-column align-items-center flex-grow-1 p-3">
                            <ChatMessagesIcon />
                            <div className="mb-2 mt-3">
                                {subscription.applyProRateLimits ? (
                                    <Text weight="bold" className={classNames('d-inline mb-0', styles.counter)}>
                                        Unlimited
                                    </Text>
                                ) : usageData?.currentUser ? (
                                    <>
                                        <Text
                                            weight="bold"
                                            className={classNames(
                                                'd-inline mb-0',
                                                styles.counter,
                                                chatLimitReached ? 'text-danger' : 'text-muted'
                                            )}
                                        >
                                            {Math.min(codyCurrentPeriodChatUsage, codyCurrentPeriodChatLimit)} /
                                        </Text>{' '}
                                        <Text
                                            className={classNames(
                                                'd-inline b-0',
                                                chatLimitReached ? 'text-danger' : 'text-muted'
                                            )}
                                            size="small"
                                        >
                                            {codyCurrentPeriodChatLimit}
                                        </Text>
                                    </>
                                ) : (
                                    <LoadingSpinner />
                                )}
                            </div>
                            <H4 className={classNames('mb-0', chatLimitReached ? 'text-danger' : 'text-muted')}>
                                Chat messages and commands
                            </H4>
                            {!subscription.applyProRateLimits &&
                                (chatLimitReached && subscription.currentPeriodEndAt ? (
                                    <Text className="text-danger mb-0" size="small">
                                        Renews <Timestamp date={usageRefreshTime} />
                                    </Text>
                                ) : (
                                    <Text className="text-muted mb-0" size="small">
                                        this month
                                    </Text>
                                ))}
                        </div>
                        {!hasTrialEnded && userIsOnProTier && hasNotAddedCreditCard && (
                            <div className="d-flex flex-column align-items-center flex-grow-1 p-3 border-left">
                                <TrialPeriodIcon />
                                <div className="mb-2 mt-4">
                                    <Text weight="bold" className={classNames('d-inline mb-0', styles.counter)}>
                                        Free trial
                                    </Text>
                                </div>
                                <Text className="text-muted mb-0" size="small">
                                    {freeTrialEndString}
                                </Text>
                            </div>
                        )}
                    </div>
                </div>

                <div className={classNames('p-4 border bg-1 mt-4 mb-5', styles.container)}>
                    <div className="d-flex justify-content-between align-items-center border-bottom pb-3">
                        <div>
                            <H2>Use Cody directly in your editor</H2>
                            <Text className="text-muted mb-0">
                                Download the Cody extension in your editor to start using Cody.
                            </Text>
                        </div>
                        <div>
                            <Link
                                to="https://sourcegraph.com/community"
                                target="_blank"
                                rel="noreferrer"
                                className="text-muted text-sm"
                            >
                                <Icon svgPath={mdiHelpCircleOutline} className="mr-1" aria-hidden={true} />
                                Have feedback? Join our community Discord to let us know!
                            </Link>
                        </div>
                    </div>
                    {editorGroups.map((group, index) => (
                        <div
                            key={group.map(editor => editor.name).join('-')}
                            className={classNames('d-flex mt-3', styles.responsiveContainer, {
                                'border-bottom pb-3': index < group.length - 1,
                            })}
                        >
                            {group.map((editor, index) => (
                                <div
                                    key={editor.name}
                                    className={classNames('d-flex flex-column flex-1 pt-3 px-3', {
                                        'border-left': index !== 0,
                                    })}
                                >
                                    <div
                                        className={classNames('d-flex mb-3 align-items-center', styles.ideHeader)}
                                        onClick={() => {
                                            setSelectedEditor(editor)
                                            setSelectedEditorStep(0)
                                        }}
                                        role="button"
                                        tabIndex={0}
                                        onKeyDown={e => {
                                            if (e.key === 'Enter') {
                                                setSelectedEditor(editor)
                                                setSelectedEditorStep(0)
                                            }
                                        }}
                                    >
                                        <div>
                                            <img
                                                alt={editor.name}
                                                src={`https://storage.googleapis.com/sourcegraph-assets/ideIcons/ideIcon${editor.icon}.svg`}
                                                width={34}
                                                className="mr-3"
                                            />
                                        </div>
                                        <div>
                                            <Text className="text-muted mb-0" size="small">
                                                {editor.publisher}
                                            </Text>
                                            <Text className={classNames('mb-0', styles.ideName)}>{editor.name}</Text>
                                            <H5 className={styles.releaseStage}>{editor.releaseStage}</H5>
                                        </div>
                                    </div>

                                    {editor.instructions && (
                                        <Link
                                            to="#"
                                            className="mb-2 text-muted d-flex align-items-center"
                                            onClick={() => {
                                                setSelectedEditor(editor)
                                                setSelectedEditorStep(0)
                                            }}
                                        >
                                            <Icon svgPath={mdiInformationOutline} aria-hidden={true} className="mr-1" />{' '}
                                            Quickstart guide
                                        </Link>
                                    )}
                                    {editor.docs && (
                                        <Link
                                            to={editor.docs}
                                            target="_blank"
                                            rel="noopener"
                                            className="text-muted d-flex align-items-center"
                                        >
                                            <Icon svgPath={mdiOpenInNew} aria-hidden={true} className="mr-1" />{' '}
                                            Documentation
                                        </Link>
                                    )}
                                    {selectedEditor?.name === editor.name &&
                                        selectedEditorStep !== null &&
                                        editor.instructions && (
                                            <Modal
                                                key={editor.name + '-modal'}
                                                isOpen={true}
                                                aria-label={`${editor.name} Info`}
                                                className={styles.modal}
                                                position="center"
                                            >
                                                <editor.instructions
                                                    showStep={selectedEditorStep}
                                                    onClose={() => {
                                                        setSelectedEditor(null)
                                                        setSelectedEditorStep(null)
                                                    }}
                                                />
                                            </Modal>
                                        )}
                                </div>
                            ))}
                            {group.length < 4
                                ? [...new Array(4 - group.length)].map((_, index) => (
                                      // eslint-disable-next-line react/no-array-index-key
                                      <div key={index} className="flex-1 p-3" />
                                  ))
                                : null}
                        </div>
                    ))}
                </div>
            </Page>
            <CodyOnboarding authenticatedUser={authenticatedUser} />
        </>
    )
}

const UpgradeToProBanner: React.FunctionComponent<{
    userIsOnProTier: boolean
}> = ({ userIsOnProTier }) =>
    userIsOnProTier ? null : (
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
                <ButtonLink to="/cody/subscription" variant="primary" size="sm">
                    Upgrade
                </ButtonLink>
            </div>
        </div>
    )

const DoNotLoseCodyProBanner: React.FunctionComponent<{
    userIsOnProTier: boolean
    arePaymentsEnabled: boolean
    hasTrialEnded: boolean
    subscriptionStatus: CodySubscriptionStatus
}> = ({ userIsOnProTier, arePaymentsEnabled, hasTrialEnded, subscriptionStatus }) => {
    const codyPaymentsUrl = useCodyPaymentsUrl()
    const manageSubscriptionRedirectURL = `${codyPaymentsUrl}/cody/subscription`

    if (arePaymentsEnabled && userIsOnProTier && subscriptionStatus === CodySubscriptionStatus.PENDING) {
        return (
            <div
                className={classNames(
                    'd-flex justify-content-between align-items-center p-4',
                    styles.dontLoseCodyProBanner
                )}
            >
                <div className="d-flex align-items-center text-dark">
                    <div className={styles.creditCardEmoji}>💳</div>
                    <div className="ml-3">
                        <H3>Keep using Cody Pro</H3>
                        <Text className="mb-0">
                            {hasTrialEnded ? (
                                <span>Enter your credit card details now and keep your Pro subscription.</span>
                            ) : (
                                <span>
                                    Enter your credit card details now and keep your subscription after your trial ends.
                                    You will only be charged on <strong>Feb 21, 2024</strong>.
                                </span>
                            )}
                        </Text>
                    </div>
                </div>
                <div>
                    <ButtonLink to={manageSubscriptionRedirectURL} variant="primary" size="sm">
                        <Icon svgPath={mdiCreditCardOutline} className="mr-1" aria-hidden={true} />
                        Add Credit Card
                    </ButtonLink>
                </div>
            </div>
        )
    }

    return null
}
