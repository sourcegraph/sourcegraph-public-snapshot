import React, { useMemo } from 'react'

import { EmbeddedCheckoutProvider, EmbeddedCheckout } from '@stripe/react-stripe-js'
import type { Stripe } from '@stripe/stripe-js'
import { useSearchParams } from 'react-router-dom'

import { H3, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { requestSSC } from '../../../util'
import { Client } from '../../api/client'
import { useApiCaller } from '../../api/hooks/useApiClient'
import { CreateCheckoutSessionRequest } from '../../api/types'

/**
 * CodyProCheckoutForm is essentially an iframe that the Stripe Elements library will
 * render an iframe into, that will host a Stripe Checkout-hosted form.
 */
export const CodyProCheckoutForm: React.FunctionComponent<{
    stripePromise: Promise<Stripe | null>
    customerEmail: string | undefined
}> = ({ stripePromise, customerEmail }) => {
    // Optionally support the "showCouponCodeAtCheckout" URL query parameter, which, if present,
    // will display a "promotional code" element in the Stripe Checkout UI.
    const [urlSearchParams] = useSearchParams()
    const showPromoCodeField = urlSearchParams.get('showCouponCodeAtCheckout') !== null

    // Make the API call to create the Stripe Checkout session.
    const call = useMemo(() => {
        const req: CreateCheckoutSessionRequest = {
            interval: 'monthly',
            seats: 1,
            customerEmail,
            showPromoCodeField,

            // URL the user is redirected to when the checkout process is complete.
            //
            // CHECKOUT_SESSION_ID will be replaced by Stripe with the correct value,
            // when the user finishes the Stripe-hosted checkout form.
            //
            // BUG: Due to the race conditions between Stripe, the SSC backend,
            // and Sourcegraph.com, immediately loading the Dashboard page isn't
            // going to show the right data reliably. We will need to instead show
            // some prompt, to give the backends an opportunity to sync.
            returnUrl: `${origin}/cody/manage?session_id={CHECKOUT_SESSION_ID}`,
        }
        return Client.createStripeCheckoutSession(req)
    }, [customerEmail, showPromoCodeField])
    const { loading, error, data } = useApiCaller(call)

    // Show a spinner while we wait for the Checkout session to be created.
    if (loading) {
        return <LoadingSpinner />
    }

    // Error page if we aren't able to show the Checkout session.
    if (error) {
        return (
            <div>
                <H3>Awe snap!</H3>
                <Text>There was an error creating the checkout session: {error.message}</Text>
            </div>
        )
    }

    return (
        <div>
            {data && data.clientSecret && (
                <EmbeddedCheckoutProvider stripe={stripePromise} options={{ clientSecret: data.clientSecret }}>
                    <EmbeddedCheckout />
                </EmbeddedCheckoutProvider>
            )}
        </div>
    )
}
