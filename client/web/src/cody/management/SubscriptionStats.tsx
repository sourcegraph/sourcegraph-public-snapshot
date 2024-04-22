import React from 'react'

import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { type CodySubscriptionStatus, CodySubscriptionPlan } from '@sourcegraph/shared/src/graphql-operations'
import { Text, LoadingSpinner, H4 } from '@sourcegraph/wildcard'

import type { UserCodyUsageResult } from '../../graphql-operations'
import { AutocompletesIcon, ChatMessagesIcon } from '../components/CodyIcon'
import { ProTierIcon } from '../subscription/CodySubscriptionPage'

import styles from './CodyManagementPage.module.scss'

interface SubscriptionStatsProps {
    subscription: {
        status: CodySubscriptionStatus
        plan: CodySubscriptionPlan
        applyProRateLimits: boolean
        currentPeriodStartAt: string
        currentPeriodEndAt: string
        cancelAtPeriodEnd: boolean
    }
    usageData: UserCodyUsageResult | undefined
}

export const SubscriptionStats: React.FunctionComponent<SubscriptionStatsProps> = ({
    subscription,
    usageData,
}: SubscriptionStatsProps) => {
    const stats = usageData?.currentUser
    const codyCurrentPeriodChatLimit = stats?.codyCurrentPeriodChatLimit || 0
    const codyCurrentPeriodChatUsage = stats?.codyCurrentPeriodChatUsage || 0
    const codyCurrentPeriodCodeLimit = stats?.codyCurrentPeriodCodeLimit || 0
    const codyCurrentPeriodCodeUsage = stats?.codyCurrentPeriodCodeUsage || 0

    const codeLimitReached = codyCurrentPeriodCodeUsage >= codyCurrentPeriodCodeLimit && codyCurrentPeriodCodeLimit > 0
    const chatLimitReached = codyCurrentPeriodChatUsage >= codyCurrentPeriodChatLimit && codyCurrentPeriodChatLimit > 0
    const isUserOnProTier = subscription.plan === CodySubscriptionPlan.PRO

    // Flag usage limits as resetting based on the current subscription's billing cycle.
    //
    // BUG: The usage limit refresh should be independent of a user's subscription data.
    //      e.g. if we offered an annual billing plan, we'd want to reset usage more often.
    //      sourcegraph#59990 is related, and required for the times to line up with the
    //      behavior from Cody Gateway.
    //
    // BUG: If the subscription is canceled, this will be in the past and therefore invalid.
    //      This data should be fetched from the SSC backend, and like above, separate
    //      from the user's subscription billing cycle.
    const usageRefreshTime = subscription.currentPeriodEndAt

    // Time when the user's current subscription will end.
    //
    // BUG: If the subscription is in the canceled state, this will be in the past. We need
    //      to update the UI to simply say "subscription canceled" or "you are on the free"
    //      plan, you don't have any subscription billing cycle anchors".
    //
    const codyProSubscriptionEndTime = subscription.currentPeriodEndAt

    return (
        <div className={classNames('d-flex align-items-center mt-3', styles.responsiveContainer)}>
            <div className="d-flex flex-column align-items-center flex-grow-1 p-3">
                {isUserOnProTier ? <ProTierIcon /> : <Text className={classNames(styles.planName, 'mb-0')}>Free</Text>}
                <Text className="text-muted mb-0" size="small">
                    tier
                </Text>
                {isUserOnProTier && subscription.cancelAtPeriodEnd && (
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
                                className={classNames('d-inline b-0', codeLimitReached ? 'text-danger' : 'text-muted')}
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
                                className={classNames('d-inline b-0', chatLimitReached ? 'text-danger' : 'text-muted')}
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
                    (chatLimitReached && usageRefreshTime ? (
                        <Text className="text-danger mb-0" size="small">
                            Renews <Timestamp date={usageRefreshTime} />
                        </Text>
                    ) : (
                        <Text className="text-muted mb-0" size="small">
                            this month
                        </Text>
                    ))}
            </div>
        </div>
    )
}
