import React from 'react'

import { Text } from '@sourcegraph/wildcard'

import type { Subscription } from '../../api/teamSubscriptions'

export const NonEditableBillingAddress: React.FC<{ subscription: Subscription }> = ({
    subscription: { name, address },
}) => (
    <div>
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
