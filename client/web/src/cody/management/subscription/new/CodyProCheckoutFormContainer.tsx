import React, { useMemo } from 'react'

import { CustomCheckoutProvider } from '@stripe/react-stripe-js'
import type { Stripe } from '@stripe/stripe-js'
import { useSearchParams } from 'react-router-dom'

import { useTheme, Theme } from '@sourcegraph/shared/src/theme'
import { H3, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { Client } from '../../api/client'
import { useApiCaller } from '../../api/hooks/useApiClient'
import type { CreateCheckoutSessionRequest } from '../../api/types'

import { CodyProCheckoutForm } from './CodyProCheckoutForm'

export const CodyProCheckoutFormContainer: React.FunctionComponent<{
    stripe: Stripe | null
    initialSeatCount: number
    customerEmail: string | undefined
}> = ({ stripe, initialSeatCount, customerEmail }) => {
    const [urlSearchParams] = useSearchParams()
    // Optionally support the "showCouponCodeAtCheckout" URL query parameter, which, if present,
    // will display a "promotional code" element in the Stripe Checkout UI.
    const showPromoCodeField = urlSearchParams.get('showCouponCodeAtCheckout') !== null

    const { theme } = useTheme()

    // Make the API call to create the Stripe Checkout session.
    const createStripeCheckoutSessionCall = useMemo(() => {
        const requestBody: CreateCheckoutSessionRequest = {
            interval: 'monthly',
            seats: initialSeatCount,
            canChangeSeatCount: true, // Seat count is always adjustable.
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
            returnUrl: `${origin}/cody/manage?session_id={CHECKOUT_SESSION_ID}&welcome=1`,
            stripeUiMode: 'custom',
        }
        return Client.createStripeCheckoutSession(requestBody)
    }, [customerEmail, initialSeatCount, showPromoCodeField])
    const { loading, error, data } = useApiCaller(createStripeCheckoutSessionCall)

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
            {data?.clientSecret && (
                <CustomCheckoutProvider
                    stripe={stripe}
                    options={{
                        clientSecret: data.clientSecret,
                        elementsOptions: { appearance: { theme: theme === Theme.Dark ? 'night' : 'stripe' } },
                    }}
                >
                    <CodyProCheckoutForm creatingTeam={initialSeatCount > 1} customerEmail={customerEmail} />
                </CustomCheckoutProvider>
            )}
        </div>
    )
}
