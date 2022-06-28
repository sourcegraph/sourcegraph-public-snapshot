import { useMemo } from 'react'

import { Elements } from '@stripe/react-stripe-js'
import { loadStripe } from '@stripe/stripe-js'

import { billingPublishableKey } from '../productSubscriptions/features'

/**
 * Wraps a React tree (of elements) and injects the Stripe API.
 */
export const StripeWrapper: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => {
    const stripe = useMemo(() => (billingPublishableKey ? loadStripe(billingPublishableKey) : null), [])

    if (stripe === null) {
        throw new Error('Stripe is not configured')
    }

    return <Elements stripe={stripe}>{children}</Elements>
}
