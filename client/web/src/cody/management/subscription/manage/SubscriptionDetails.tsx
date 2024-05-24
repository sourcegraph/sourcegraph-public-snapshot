import { useContext, useEffect, useState } from 'react'

import { mdiCancel, mdiCheck, mdiRefresh } from '@mdi/js'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { Button, H1, H3, Icon, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

import { Client } from '../../api/client'
import { CodyProApiClientContext } from '../../api/components/CodyProApiClient'
import type { Subscription } from '../../api/teamSubscriptions'

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
    refetchSubscription: () => Promise<void>
}

export const SubscriptionDetails: React.FC<SubscriptionDetailsProps> = props => {
    const { caller } = useContext(CodyProApiClientContext)

    const [isLoading, setIsLoading] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')

    useEffect(
        function clearErrorMessageAfterTimeout() {
            let timeout: NodeJS.Timeout
            if (errorMessage) {
                timeout = setTimeout(() => setErrorMessage(''), 5000)
            }
            return () => {
                if (timeout) {
                    clearTimeout(timeout)
                }
            }
        },
        [errorMessage]
    )

    const updateSubscription = async (type: 'cancel' | 'renew'): Promise<void> => {
        setIsLoading(true)
        setErrorMessage('')
        const serverErrorText = `An error occurred while ${type}ing the subscription. Please try again. If the problem persists, contact support at support@sourcegraph.com.`
        try {
            const { response } = await caller.call(
                Client.updateCurrentSubscription({
                    subscriptionUpdate: { newCancelAtPeriodEnd: type === 'cancel' },
                })
            )

            if (response.ok) {
                await props.refetchSubscription()
            } else {
                setErrorMessage(serverErrorText)
            }
        } catch (error) {
            logger.error(error)
            setErrorMessage(serverErrorText)
        } finally {
            setIsLoading(false)
        }
    }

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
                        <Text as="span" className={styles.bold}>
                            {humanizeDate(props.subscription.currentPeriodEnd)}
                        </Text>
                        .
                    </Text>
                </div>
                {props.subscription.cancelAtPeriodEnd ? (
                    <RenewSubscriptionButton isLoading={isLoading} onClick={() => updateSubscription('renew')} />
                ) : (
                    <CancelSubscriptionButton
                        isLoading={isLoading}
                        currentPeriodEnd={props.subscription.currentPeriodEnd}
                        onClick={() => updateSubscription('cancel')}
                    />
                )}
            </div>

            {errorMessage && <Text className="mt-3 text-danger">{errorMessage}</Text>}
        </>
    )
}

const RenewSubscriptionButton: React.FC<{
    isLoading: boolean
    onClick: () => Promise<void>
}> = props => (
    <Button variant="primary" disabled={props.isLoading} className={styles.iconButton} onClick={props.onClick}>
        {props.isLoading ? (
            <LoadingSpinner className="mr-1" />
        ) : (
            <Icon aria-hidden={true} className="mr-1" svgPath={mdiRefresh} />
        )}
        Renew subscription
    </Button>
)

const CancelSubscriptionButton: React.FC<{
    isLoading: boolean
    currentPeriodEnd: Subscription['currentPeriodEnd']
    onClick: () => Promise<void>
}> = props => {
    const [isConfirmationModalVisible, setIsConfirmationModalVisible] = useState(false)

    return (
        <>
            <Button variant="secondary" onClick={() => setIsConfirmationModalVisible(true)}>
                <Icon aria-hidden={true} svgPath={mdiCancel} className="mr-1" />
                Cancel subscription
            </Button>

            {isConfirmationModalVisible && (
                <Modal aria-label="Confirmation modal" onDismiss={() => setIsConfirmationModalVisible(false)}>
                    <div className="pb-3">
                        <H3>Are you sure?</H3>
                        <Text className="mt-4">
                            Canceling you subscription now means that you won't be able to use Cody with Pro features
                            after {humanizeDate(props.currentPeriodEnd)}.
                        </Text>
                        <Text className={classNames('mt-4 mb-0', styles.bold)}>Do you want to procceed?</Text>
                    </div>
                    <div className={classNames('d-flex mt-4', styles.buttonContainer)}>
                        <Button variant="secondary" outline={true} onClick={() => setIsConfirmationModalVisible(false)}>
                            No, I've changed my mind
                        </Button>
                        <Button
                            variant="primary"
                            disabled={props.isLoading}
                            className={styles.iconButton}
                            onClick={() => props.onClick().finally(() => setIsConfirmationModalVisible(false))}
                        >
                            {props.isLoading ? (
                                <LoadingSpinner className="mr-1" />
                            ) : (
                                <Icon aria-hidden={true} className="mr-1" svgPath={mdiCheck} />
                            )}
                            <Text as="span">Yes, cancel</Text>
                        </Button>
                    </div>
                </Modal>
            )}
        </>
    )
}
