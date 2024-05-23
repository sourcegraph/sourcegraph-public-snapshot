import { useEffect, useState } from 'react'

import { mdiCancel } from '@mdi/js'
import classNames from 'classnames'

import { Button, H1, H3, Icon, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

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

export const SubscriptionDetails: React.FC<{ subscription: Subscription }> = ({ subscription }) => {
    const [errorMessage, setErrorMessage] = useState('')
    const [confirmCancelModalVisible, setConfirmCancelModalVisible] = useState(false)
    const [changingCancellationStatus, setChangingCancellationStatus] = useState(false)

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

    return (
        <>
            <div className={classNames('d-flex', styles.container)}>
                <div>
                    {subscription.nextInvoice && (
                        <Text className="mb-0">
                            <H1 as="span">{usdCentsToHumanString(subscription.nextInvoice.newPrice)}</H1>
                            <Text as="span" className={classNames('text-muted', styles.period)}>
                                {' '}
                                /{formatBillingInterval(subscription.billingInterval)}
                            </Text>
                        </Text>
                    )}
                    <Text className="mb-0">
                        Subscription{' '}
                        {subscription.cancelAtPeriodEnd ? 'canceled. Access to Cody Pro will end' : 'renews'} on{' '}
                        <Text as="span" className={styles.bold}>
                            {humanizeDate(subscription.currentPeriodEnd)}
                        </Text>
                        .
                    </Text>
                </div>
                <Button variant="secondary" onClick={() => setConfirmCancelModalVisible(true)}>
                    {changingCancellationStatus && <LoadingSpinner />}
                    {subscription.cancelAtPeriodEnd ? (
                        'Resume Subscription'
                    ) : (
                        <span className="inline-flex items-center">
                            <Icon aria-hidden={true} svgPath={mdiCancel} className="mr-1" />
                            Cancel subscription
                        </span>
                    )}
                </Button>
            </div>

            {errorMessage && <Text className="mt-3 text-danger">{errorMessage}</Text>}

            {confirmCancelModalVisible && (
                <Modal aria-label="Confirmation modal" onDismiss={() => setConfirmCancelModalVisible(false)}>
                    <div className="pb-3">
                        <H3>Are you sure?</H3>
                        <Text className="mt-4">
                            Canceling you subscription now means that you won't be able to use Cody with Pro features
                            after {humanizeDate(subscription.currentPeriodEnd)}.
                        </Text>
                        <Text className={classNames('mt-4 mb-0', styles.bold)}>Do you want to procceed?</Text>
                    </div>
                    <div className={classNames('d-flex mt-4', styles.buttonContainer)}>
                        <Button variant="secondary" outline={true} onClick={() => setConfirmCancelModalVisible(false)}>
                            No, I've changed my mind
                        </Button>
                        <Button variant="primary" onClick={() => setConfirmCancelModalVisible(false)}>
                            Yes, cancel
                        </Button>
                    </div>
                </Modal>
            )}
        </>
    )
}
