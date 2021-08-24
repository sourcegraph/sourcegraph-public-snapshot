import { Elements } from '@stripe/react-stripe-js'
import { loadStripe } from '@stripe/stripe-js'
import React, { useMemo } from 'react'

import { billingPublishableKey } from '../productSubscriptions/features'

type Props<P> = P & { component: React.ComponentType<Omit<P, 'component'>> }

/**
 * Wraps a React tree (of elements) and injects the Stripe API.
 */
export const StripeWrapper = <P extends object>({ component: Component, ...props }: Props<P>): JSX.Element | null => {
    const stripe = useMemo(() => (billingPublishableKey ? loadStripe(billingPublishableKey) : null), [])
    if (stripe === null) {
        throw new Error('Stripe is not configured')
    }
    return (
        <Elements stripe={stripe}>
            <Component {...props} />
        </Elements>
    )
}
