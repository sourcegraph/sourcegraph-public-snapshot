import React, { useEffect, useState } from 'react'

import { mdiCancel, mdiCheck, mdiRefresh } from '@mdi/js'
import classNames from 'classnames'

import { Button, H1, H3, Icon, Modal, Text } from '@sourcegraph/wildcard'

import { useUpdateCurrentSubscription } from '../../api/react-query/subscriptions'
import type { Subscription } from '../../api/teamSubscriptions'

import { LoadingIconButton } from './LoadingIconButton'
import { humanizeDate, usdCentsToHumanString } from './utils'

import styles from './SubscriptionDetails.module.scss'

enum BillingInterval {
    Daily = 'DAILY',
    Monthly = 'MONTHLY',
    Yearly = 'YEARLY',
}

const formatBillingInterval = (interval: string): string =>
    ({
        [BillingInterval.Daily]: 'day',
        [BillingInterval.Monthly]: 'month',
        [BillingInterval.Yearly]: 'year',
    }[interval] ?? 'period')

interface SubscriptionDetailsProps {
    subscription: Subscription
}

export const SubscriptionDetails: React.FC<SubscriptionDetailsProps> = props => {
    const updateCurrentSubscriptionMutation = useUpdateCurrentSubscription()

    const [isConfirmationModalVisible, setIsConfirmationModalVisible] = useState(false)
    const [isErrorVisible, setIsErrorVisible] = useState(false)

    const errorMessage = updateCurrentSubscriptionMutation.isError
        ? 'An error occurred while updating your subscription status. Please try again. If the problem persists, contact support at support@sourcegraph.com.'
        : ''

    useEffect(
        function clearErrorMessageAfterTimeout() {
            let timeout: NodeJS.Timeout
            if (errorMessage) {
                setIsErrorVisible(true)
                timeout = setTimeout(() => setIsErrorVisible(false), 5000)
            }
            return () => {
                if (timeout) {
                    clearTimeout(timeout)
                }
            }
        },
        [errorMessage]
    )

    return (
        <>
            <div className={classNames('d-flex', styles.container)}>
                <div>
                    {props.subscription.nextInvoice && (
                        <Text className="mb-0">
                            <H1 as="span">{usdCentsToHumanString(props.subscription.nextInvoice.newPrice)}</H1>
                            <Text as="span" className={classNames('text-muted', styles.period)}>
                                {' '}
                                /{formatBillingInterval(props.subscription.billingInterval)}
                            </Text>
                        </Text>
                    )}
                    <Text className="mb-0">
                        {props.subscription.cancelAtPeriodEnd
                            ? 'Subscription canceled. Access to Cody Pro will end on'
                            : 'Subscription renews on'}{' '}
                        <Text as="span" className="font-bold">
                            {humanizeDate(props.subscription.currentPeriodEnd)}
                        </Text>
                        .
                    </Text>
                </div>
                {props.subscription.cancelAtPeriodEnd ? (
                    <LoadingIconButton
                        variant="primary"
                        disabled={updateCurrentSubscriptionMutation.isPending}
                        isLoading={updateCurrentSubscriptionMutation.isPending}
                        onClick={() =>
                            updateCurrentSubscriptionMutation.mutate({
                                subscriptionUpdate: { newCancelAtPeriodEnd: false },
                            })
                        }
                        iconSvgPath={mdiRefresh}
                    >
                        Renew subscription
                    </LoadingIconButton>
                ) : (
                    <>
                        <Button variant="secondary" outline={true} onClick={() => setIsConfirmationModalVisible(true)}>
                            <Icon aria-hidden={true} svgPath={mdiCancel} className="mr-1" />
                            Cancel subscription
                        </Button>

                        {isConfirmationModalVisible && (
                            <Modal
                                aria-label="Confirmation modal"
                                onDismiss={() => setIsConfirmationModalVisible(false)}
                            >
                                <div className="pb-3">
                                    <H3>Are you sure?</H3>
                                    <Text className="mt-4">
                                        Canceling your subscription now means that you won't be able to use Cody with
                                        Pro features after {humanizeDate(props.subscription.currentPeriodEnd)}.
                                    </Text>
                                    <Text className="mt-4 mb-0 font-bold">Do you want to proceed?</Text>
                                </div>
                                <div className="d-flex mt-4 justify-content-end">
                                    <Button
                                        variant="secondary"
                                        outline={true}
                                        onClick={() => setIsConfirmationModalVisible(false)}
                                        className="mr-3"
                                    >
                                        No, I've changed my mind
                                    </Button>
                                    <LoadingIconButton
                                        variant="primary"
                                        disabled={updateCurrentSubscriptionMutation.isPending}
                                        isLoading={updateCurrentSubscriptionMutation.isPending}
                                        onClick={() =>
                                            updateCurrentSubscriptionMutation.mutate(
                                                { subscriptionUpdate: { newCancelAtPeriodEnd: true } },
                                                { onSettled: () => setIsConfirmationModalVisible(false) }
                                            )
                                        }
                                        iconSvgPath={mdiCheck}
                                    >
                                        Yes, cancel
                                    </LoadingIconButton>
                                </div>
                            </Modal>
                        )}
                    </>
                )}
            </div>

            {updateCurrentSubscriptionMutation.isError && isErrorVisible ? (
                <Text className="mt-3 text-danger">{errorMessage}</Text>
            ) : null}
        </>
    )
}
