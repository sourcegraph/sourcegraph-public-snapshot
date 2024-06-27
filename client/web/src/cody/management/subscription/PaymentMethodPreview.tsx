import React from 'react'

import { mdiPencilOutline, mdiCreditCardOutline, mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { H3, Button, Icon, Text } from '@sourcegraph/wildcard'

import type { Subscription } from '../api/teamSubscriptions'

import styles from './manage/PaymentDetails.module.scss'

export const PaymentMethodPreview: React.FC<
    Pick<Subscription, 'paymentMethod'> & { isEditable: boolean; onButtonClick?: () => void; className?: string }
> = ({ paymentMethod, isEditable, onButtonClick = () => undefined, className }) =>
    paymentMethod ? (
        <div className={className}>
            <div className="d-flex align-items-center justify-content-between">
                <H3>Active credit card</H3>
                {isEditable && (
                    <Button variant="link" className={styles.titleButton} onClick={onButtonClick}>
                        <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
                    </Button>
                )}
            </div>
            <div className="mt-3 d-flex justify-content-between">
                <Text as="span" className={classNames('text-muted', styles.paymentMethodNumber)}>
                    <Icon aria-hidden={true} svgPath={mdiCreditCardOutline} /> ···· ···· ···· {paymentMethod.last4}
                </Text>
                <Text as="span" className="text-muted">
                    Expires {paymentMethod.expMonth}/{paymentMethod.expYear}
                </Text>
            </div>
        </div>
    ) : (
        <div className={classNames('d-flex align-items-center justify-content-between', className)}>
            <H3>No payment method is available</H3>
            {isEditable && (
                <Button variant="link" className={styles.titleButton} onClick={onButtonClick}>
                    <Icon aria-hidden={true} svgPath={mdiPlus} className="mr-1" /> Add
                </Button>
            )}
        </div>
    )
