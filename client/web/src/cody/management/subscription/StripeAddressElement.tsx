import React, { useMemo } from 'react'

import { AddressElement } from '@stripe/react-stripe-js'
import type { StripeAddressElementOptions } from '@stripe/stripe-js'

import type { Subscription } from '../api/types'

interface StripeAddressElementProps {
    subscription?: Subscription
    onFocus?: () => void
}

export const StripeAddressElement: React.FC<StripeAddressElementProps> = ({ subscription, onFocus }) => {
    const options = useMemo(
        (): StripeAddressElementOptions => ({
            mode: 'billing',
            display: { name: 'full' },
            ...(subscription
                ? {
                      defaultValues: {
                          name: subscription.name,
                          address: {
                              ...subscription.address,
                              postal_code: subscription.address.postalCode,
                          },
                      },
                  }
                : {}),
            autocomplete: { mode: 'automatic' },
        }),
        [subscription]
    )

    return <AddressElement options={options} onFocus={onFocus} />
}
