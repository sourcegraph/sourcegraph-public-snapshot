import React, { useMemo, useEffect } from 'react'

import { mdiMinusThick, mdiPlusThick } from '@mdi/js'
import { PaymentElement, AddressElement, useCustomCheckout, Elements } from '@stripe/react-stripe-js'
import type { Stripe } from '@stripe/stripe-js'
import classNames from 'classnames'
import { useSearchParams, useNavigate } from 'react-router-dom'

import { pluralize } from '@sourcegraph/common'
import { useTheme, Theme } from '@sourcegraph/shared/src/theme'
import {
    Button,
    Container,
    Form,
    Grid,
    H2,
    H3,
    Icon,
    Input,
    Label,
    Link,
    LoadingSpinner,
    Text,
    useDebounce,
} from '@sourcegraph/wildcard'

import { CodyAlert } from '../../../components/CodyAlert'
import { Client } from '../../api/client'
import { useApiCaller } from '../../api/hooks/useApiClient'
import type { CreatePaymentSessionRequest } from '../../api/types'

import { PayButton } from './PayButton'

import styles from './NewCodyProSubscriptionPage.module.scss'

const MIN_SEAT_COUNT = 1
const MAX_SEAT_COUNT = 50

export const CodyProCheckoutForm: React.FunctionComponent<{
    stripe: Stripe | null
    initialSeatCount: number
    customerEmail: string | undefined
}> = ({ stripe, initialSeatCount, customerEmail }) => {
    // Optionally support the "showCouponCodeAtCheckout" URL query parameter, which, if present,
    // will display a "promotional code" element in the Stripe Checkout UI.
    const [urlSearchParams] = useSearchParams()
    const showPromoCodeField = urlSearchParams.get('showCouponCodeAtCheckout') !== null

    const navigate = useNavigate()

    const { total, lineItems, updateLineItemQuantity, email, updateEmail, status } = useCustomCheckout()

    const [errorMessage, setErrorMessage] = React.useState<string | null>(null)
    const [updatingSeatCount, setUpdatingSeatCount] = React.useState(false)
    const [seatCount, setSeatCount] = React.useState(lineItems[0]?.quantity)
    const debouncedSeatCount = useDebounce(seatCount, 800)
    const firstLineItemId = lineItems[0]?.id
    const creatingTeam = initialSeatCount > 1

    useEffect(() => {
        const updateSeatCount = async (): Promise<void> => {
            setUpdatingSeatCount(true)
            try {
                await updateLineItemQuantity({
                    lineItem: firstLineItemId,
                    quantity: debouncedSeatCount,
                })
            } catch {
                setErrorMessage('Failed to update seat count. Please change the number of seats to try again.')
            }
            setUpdatingSeatCount(false)
        }

        void updateSeatCount()
    }, [firstLineItemId, debouncedSeatCount, updateLineItemQuantity])

    const isPriceLoading = seatCount !== debouncedSeatCount || updatingSeatCount

    // Set initial seat count.
    useEffect(() => {
        if (lineItems.length === 1) {
            setSeatCount(lineItems[0].quantity)
        }
    }, [lineItems])

    // Set customer email to initial value.
    useEffect(() => {
        if (customerEmail) {
            updateEmail(customerEmail)
        }
    }, [customerEmail, updateEmail])

    // Redirect once we're done.
    // Display an error message if the session is expired.
    useEffect(() => {
        if (status.type === 'complete') {
            navigate('/cody/manage?welcome=1')
        } else if (status.type === 'expired') {
            setErrorMessage('Session expired. Please refresh the page.')
        }
    }, [navigate, status.type])

    const { theme } = useTheme()

    // Make the API call to create the Stripe Payment session.
    const createStripePaymentSessionCall = useMemo(() => {
        const requestBody: CreatePaymentSessionRequest = {
            interval: 'monthly',
            seats: initialSeatCount,
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
        }
        return Client.createStripePaymentSession(requestBody)
    }, [customerEmail, initialSeatCount, showPromoCodeField])
    const { loading, error, data } = useApiCaller(createStripePaymentSessionCall)

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
            {data?.clientSecret && seatCount >= 30 && (
                <CodyAlert variant="purple">
                    <H3>Explore an enterprise plan</H3>
                    <Text className="mb-0">
                        Team plans are limited to 50 users.{' '}
                        <Link to="https://sourcegraph.com/contact/sales/">Contact sales</Link> to learn more.
                    </Text>
                </CodyAlert>
            )}
            <Container>
                <Grid columnCount={2} spacing={4}>
                    <div>
                        <H2>{creatingTeam ? 'Add seats' : 'Select number of seats'}</H2>
                        <div className="d-flex flex-row align-items-center pb-3 mb-4 border-bottom">
                            <div className="flex-1">$9 per seat / month</div>
                            <Button
                                disabled={seatCount === MIN_SEAT_COUNT}
                                onClick={() => setSeatCount(c => (c > MIN_SEAT_COUNT ? c - 1 : c))}
                            >
                                <Icon aria-hidden={true} svgPath={mdiMinusThick} />
                            </Button>
                            <div className={styles.seatCountSelectorValue}>{seatCount}</div>
                            <Button
                                disabled={seatCount === MAX_SEAT_COUNT}
                                onClick={() => setSeatCount(c => (c < MAX_SEAT_COUNT ? c + 1 : c))}
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlusThick} />
                            </Button>
                        </div>
                        <H2>Summary</H2>
                        <div className="d-flex flex-row align-items-center mb-4">
                            <div className="flex-1">
                                {creatingTeam ? 'Adding ' : ''} {seatCount} {pluralize('seat', seatCount)}
                            </div>
                            <div>
                                <strong>
                                    {isPriceLoading ? (
                                        <LoadingSpinner className={styles.lineHeightLoadingSpinner} />
                                    ) : (
                                        `$${total.total / 100} / month`
                                    )}
                                </strong>
                            </div>
                        </div>
                        <Text size="small">
                            <em>Each seat is pro-rated this month, and will be charged at the full rate next month.</em>
                        </Text>
                    </div>
                    <div>
                        <H2>
                            Purchase {seatCount} {pluralize('seat', seatCount)}
                        </H2>
                        <Label>Email</Label>
                        <Input value={email || ''} disabled={true} className="mb-4" />
                        <Form>
                            <Elements
                                stripe={stripe}
                                options={{
                                    clientSecret: data?.clientSecret,
                                    appearance: { theme: theme === Theme.Dark ? 'night' : 'stripe' },
                                }}
                            >
                                <PaymentElement options={{ layout: 'accordion' }} className="mb-4" />
                                <AddressElement options={{ mode: 'billing' }} />
                            </Elements>
                            {errorMessage && (
                                <div className={classNames(styles.paymentDataErrorMessage)}>{errorMessage}</div>
                            )}

                            <PayButton
                                setErrorMessage={setErrorMessage}
                                className={classNames('d-block w-100 mb-4', styles.payButton)}
                            >
                                Subscribe
                            </PayButton>
                            <div>
                                <Text>
                                    By clicking the button, you agree to the{' '}
                                    <Link to="/terms/cloud">Terms of Service</Link> and acknowledge that the{' '}
                                    <Link to="/terms/privacy">Privacy Statement</Link> applies. Your subscription will
                                    renew automatically by charging your payment method on file until you{' '}
                                    <Link to="/docs/cody/usage-and-pricing#downgrading-from-pro-to-free">cancel</Link>.
                                    You may cancel at any time prior to the next billing cycle.
                                </Text>
                            </div>
                        </Form>
                    </div>
                </Grid>
            </Container>
        </div>
    )
}
