import React from 'react'

import { mdiPencilOutline } from '@mdi/js'

import { Text, H3, Button, Icon } from '@sourcegraph/wildcard'

import type { Subscription } from '../api/teamSubscriptions'

import styles from './manage/PaymentDetails.module.scss'

export const BillingAddressPreview: React.FC<{
    subscription: Subscription
    isEditable: boolean
    onButtonClick?: () => void
    className?: string
}> = ({ subscription: { name, address }, isEditable, onButtonClick = () => undefined, className }) => (
    <div className={className}>
        <div className="d-flex align-items-center justify-content-between">
            <H3>Billing address</H3>
            {isEditable && (
                <Button variant="link" className={styles.titleButton} onClick={onButtonClick}>
                    <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
                </Button>
            )}
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Full name
            </Text>
            <Text className="font-weight-medium">{name}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Country or region
            </Text>
            <Text className="font-weight-medium">{address.country || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Address line 1
            </Text>
            <Text className="font-weight-medium">{address.line1 || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Address line 2
            </Text>
            <Text className="font-weight-medium">{address.line2 || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                City
            </Text>
            <Text className="font-weight-medium">{address.city || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                State
            </Text>
            <Text className="font-weight-medium">{address.state || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Postal code
            </Text>
            <Text className="font-weight-medium">{address.postalCode || '-'}</Text>
        </div>
    </div>
)
