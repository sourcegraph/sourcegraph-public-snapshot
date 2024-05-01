import { useState, useEffect, FunctionComponent } from 'react'
import { useSearchParams } from 'react-router-dom'

import { EmbeddedCheckoutProvider, EmbeddedCheckout } from '@stripe/react-stripe-js'

import * as stripeJs from '@stripe/stripe-js'

/**
 * CodyProCheckoutForm is essentially an iframe that the Stripe Elements library will
 * render an iframe into, that will host a Stripe Checkout-hosted form.
 */
export const CodyProCheckoutForm: FunctionComponent<{
    stripeHandle: Promise<stripeJs.Stripe | null>
    customerEmail: string | undefined
}> = ({ stripeHandle, customerEmail }) => {
    const [clientSecret, setClientSecret] = useState('')
    const [errorDetails, setErrorDetails] = useState('')
    const [urlSearchParams] = useSearchParams()

    // Optionally support the "showCouponCodeAtCheckout" URL query parameter, which if present
    // will display a "promotional code" element in the Stripe Checkout UI.
    const showPromoCodeField = urlSearchParams.get('showCouponCodeAtCheckout') !== null

    // Issue an API call to the backend asking it to create a new checkout session.
    // This will update clientSecret/errorDetails asynchronously when the request completes.
    useEffect(() => {
        // useEffect will not accept a Promise, so we call
        // createCheckoutSession and let it run async.
        // (And not `await createCheckoutSession` or `return createCheckoutSession`.)
        void createCheckoutSession('monthly', showPromoCodeField, customerEmail, setClientSecret, setErrorDetails)
    }, [customerEmail, showPromoCodeField])

    const embeddedCheckoutOpts /* unexported EmbeddedCheckoutProviderProps.options */ = {
        clientSecret: clientSecret,
    }
    return (
        <div id="checkout">
            {errorDetails && (
                <div>
                    <h3>Awe snap!</h3>
                    <p>There was an error creating the checkout session: {errorDetails}</p>
                </div>
            )}

            {clientSecret && (
                <EmbeddedCheckoutProvider stripe={stripeHandle} options={embeddedCheckoutOpts}>
                    <EmbeddedCheckout />
                </EmbeddedCheckoutProvider>
            )}
        </div>
    )
}

// createSessionResponse is the API response returned from the SSC backend when
// we ask it to create a new Stripe Checkout Session.
interface createSessionResponse {
    clientSecret: string
}

// createCheckoutSession initiates the API request to the backend to create a Stripe Checkout session.
// Upon completion, the `setClientSecret` or `setErrorDetails` will be called to report the result.
async function createCheckoutSession(
    billingInterval: string,
    showPromoCodeField: boolean,
    customerEmail: string | undefined,
    setClientSecret: (arg: string) => void,
    setErrorDetails: (arg: string) => void
): Promise<void> {
    // e.g. "https://sourcegraph.com"
    const origin = window.location.origin

    try {
        // So the request is kinda made to 2x backends. dotcom's .api/ssc/proxy endpoint will
        // take care of exchanging the Sourcegraph session credentials for a SAMS access token.
        // And then proxy the request onto the SSC backend, which will actually create the
        // checkout session.
        const resp = await fetch(`${origin}/.api/ssc/proxy/checkout/session`, {
            // Pass along the "sgs" session cookie to identify the caller.
            credentials: 'same-origin',
            method: 'POST',
            // Object sent to the backend. See `createCheckoutSessionRequest`.
            body: JSON.stringify({
                interval: billingInterval,
                seats: 1,
                customerEmail: customerEmail,
                showPromoCodeField: showPromoCodeField,

                // URL the user is redirected to when the checkout process is complete.
                //
                // BUG: Due to the race conditions between Stripe, the SSC backend,
                // and Sourcegraph.com, immediately loading the Dashboard page isn't
                // going to show the right data reliably. We will need to instead show
                // some intersitular or welcome prompt, to give various things to sync.
                returnUrl: `${origin}/cody/manage?session_id={CHECKOUT_SESSION_ID}`
            }),
        })

        const respBody = await resp.text()
        if (resp.status >= 200 && resp.status <= 299) {
            const typedResp = JSON.parse(respBody) as createSessionResponse
            setClientSecret(typedResp.clientSecret)
        } else {
            // Pass any 4xx or 5xx directly to the user. We expect the
            // server to have properly redcated any sensive information.
            setErrorDetails(respBody)
        }
    } catch (ex) {
        setErrorDetails(`unhandled exception: ${JSON.stringify(ex)}`)
    }
}
