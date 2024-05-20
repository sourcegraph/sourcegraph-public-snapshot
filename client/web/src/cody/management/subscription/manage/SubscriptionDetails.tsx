import { useEffect, useState } from 'react'

import { mdiCancel } from '@mdi/js'
import classNames from 'classnames'

import { Button, H1, Icon, Text } from '@sourcegraph/wildcard'

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

    const setCancelAtPeriodEnd = async (value: 'cancel' | 'renew') => {
        setConfirmCancelModalVisible(false)
        setChangingCancellationStatus(true)
        try {
            // await client.mutate({
            //     mutation: MUTATE_TEAM_SUBSCRIPTION_CANCEL_AT_PERIOD_END,
            //     variables: {
            //         teamId: subscriptionDetails.teamId,
            //         cancelAtPeriodEnd: value === 'cancel',
            //     },
            // })

            // onChange()
            console.log('calling onChange callback')
        } catch (error) {
            // TODO[accounts.sourcegraph.com#353]: Send error to Sentry
            // eslint-disable-next-line no-console
            console.error(error)
            setErrorMessage(
                'An error occurred while canceling/resuming the subscription. Please try again. If the problem persists, contact support at support@sourcegraph.com.'
            )
        }
        setChangingCancellationStatus(false)
    }

    // Auto-hide error message.

    useEffect(() => {
        if (errorMessage) {
            setTimeout(() => {
                setErrorMessage('')
            }, 5000)
        }
    }, [errorMessage])

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
                <Button
                    variant="secondary"
                    // disabled={changingCancellationStatus}
                    onClick={() => {
                        console.log('cancelAtPeriodEnd', subscription.cancelAtPeriodEnd)
                        // if (subscriptionDetails.cancelAtPeriodEnd) {
                        //     void setCancelAtPeriodEnd('renew')
                        // } else {
                        //     setConfirmCancelModalVisible(true)
                        // }
                    }}
                >
                    {changingCancellationStatus && (
                        <div className="spinner w-4 h-4 border-2 border-gray-800 border-t-transparent rounded-full animate-spin mr-2" />
                    )}
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

            {errorMessage && <p className="bg-red-100 text-red-700 p-4 rounded">{errorMessage}</p>}

            <OkCancelModal
                isOpen={confirmCancelModalVisible}
                cancelText="No, I've changed my mind"
                okText="Yes, cancel"
                onCancel={() => {
                    setConfirmCancelModalVisible(false)
                }}
                onOk={() => {
                    void setCancelAtPeriodEnd('cancel')
                }}
            >
                <p>
                    <strong>Are you sure?</strong>
                </p>
                <p>
                    Canceling you subscription now means that you won't be able to use Cody with Pro features after{' '}
                    {/* eslint-disable-next-line @typescript-eslint/no-unsafe-argument */}
                    {humanizeDate(subscription.currentPeriodEnd)}
                </p>
                <p>
                    <strong>Do you want to proceed?</strong>
                </p>
            </OkCancelModal>
        </>
    )
}

const OkCancelModal: React.FC<{
    isOpen: boolean
    cancelText: string
    okText: string
    onCancel: () => void
    onOk: () => void
    children: React.ReactNode
}> = ({ isOpen, cancelText, onCancel, okText, onOk, children }) => {
    if (!isOpen) {
        return null
    }

    return (
        <div className="d-flex">
            <div className="">{children}</div>
            <div className="">
                <button type="button" className="" onClick={onCancel}>
                    {cancelText}
                </button>
                <button type="button" className="" onClick={onOk}>
                    {okText}
                </button>
            </div>
        </div>
    )
}
