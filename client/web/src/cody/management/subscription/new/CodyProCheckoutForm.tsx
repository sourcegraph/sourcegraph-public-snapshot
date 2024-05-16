import React from 'react'

import { EmbeddedCheckoutProvider, EmbeddedCheckout } from '@stripe/react-stripe-js'
import type { Stripe } from '@stripe/stripe-js'
import { useSearchParams } from 'react-router-dom'

import { H3, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { useApiClient, useApiCaller } from '../../api/hooks/useApiClient'
import { CreateCheckoutSessionRequest } from '../../api/types'

import { requestSSC } from '../../../util'

/**
 * CodyProCheckoutForm is essentially an iframe that the Stripe Elements library will
 * render an iframe into, that will host a Stripe Checkout-hosted form.
 */
export const CodyProCheckoutForm: React.FunctionComponent<{
    stripePromise: Promise<Stripe | null>
    customerEmail: string | undefined
}> = ({ stripePromise, customerEmail }) => {
    // Optionally support the "showCouponCodeAtCheckout" URL query parameter, which if present
    // will display a "promotional code" element in the Stripe Checkout UI.
    const [urlSearchParams] = useSearchParams()
    const showPromoCodeField = urlSearchParams.get('showCouponCodeAtCheckout') !== null

    const req: CreateCheckoutSessionRequest = {
        interval: 'monthly',
        seats: 1,
        customerEmail,
        showPromoCodeField,

        // URL the user is redirected to when the checkout process is complete.
        //
        // BUG: Due to the race conditions between Stripe, the SSC backend,
        // and Sourcegraph.com, immediately loading the Dashboard page isn't
        // going to show the right data reliably. We will need to instead show
        // some intersitular or welcome prompt, to give various things to sync.
        returnUrl: `${origin}/cody/manage?session_id={CHECKOUT_SESSION_ID}`,
    }

    // Make the API call to create the Stripe Checkout session.
    const client = useApiClient()
    const { loading, error, data } = useApiCaller(client.createStripeCheckoutSession(req))

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <div>
            {error && (
                <>
                    <H3>Awe snap!</H3>
                    <Text>There was an error creating the checkout session: {error.message}</Text>
                </>
            )}

            {data && data.clientSecret && (
                <EmbeddedCheckoutProvider stripe={stripePromise} options={{ clientSecret: data.clientSecret }}>
                    <EmbeddedCheckout />
                </EmbeddedCheckoutProvider>
            )}
        </div>
    )
}
