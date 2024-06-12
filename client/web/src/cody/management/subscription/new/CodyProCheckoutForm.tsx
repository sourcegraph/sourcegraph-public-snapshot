import React from 'react'

import { mdiMinusThick, mdiPlusThick } from '@mdi/js'
import { AddressElement, useStripe, useElements, CardNumberElement } from '@stripe/react-stripe-js'
import type { Stripe, StripeCardNumberElement } from '@stripe/stripe-js'
import type { StripeAddressElementChangeEvent } from '@stripe/stripe-js/dist/stripe-js/elements/address'
import classNames from 'classnames'
import { useNavigate, useSearchParams } from 'react-router-dom'

import { pluralize } from '@sourcegraph/common'
import { Form, Link, Button, Grid, H2, Text, Container, Icon, H3, LoadingSpinner } from '@sourcegraph/wildcard'

import { CodyAlert } from '../../../components/CodyAlert'
import { useCreateTeam } from '../../api/react-query/subscriptions'
import type { Subscription } from '../../api/types'
import { NonEditableBillingAddress } from '../manage/NonEditableBillingAddress'
import { StripeAddressElement } from '../StripeAddressElement'
import { StripeCardDetails } from '../StripeCardDetails'

import styles from './NewCodyProSubscriptionPage.module.scss'

const MIN_SEAT_COUNT = 1
const MAX_SEAT_COUNT = 50

interface CodyProCheckoutFormProps {
    subscription?: Subscription
    customerEmail: string | undefined
}

async function createStripeToken(
    stripe: Stripe,
    cardNumberElement: StripeCardNumberElement,
    suppliedAddress: StripeAddressElementChangeEvent['value']['address']
): Promise<string> {
    let response
    try {
        // Note that Stripe may have returned an error response.
        response = await stripe.createToken(cardNumberElement, {
            // We send the address data along with the card info to let Stripe do more validation such as
            // confirming the zip code matches the card's. Later, we'll also store this as the Customer's address.
            address_line1: suppliedAddress.line1,
            address_line2: suppliedAddress.line2 || '',
            address_city: suppliedAddress.city,
            address_state: suppliedAddress.state,
            address_zip: suppliedAddress.postal_code,
            address_country: suppliedAddress.country,
            currency: 'usd',
        })
    } catch (error) {
        throw new Error(`We couldn't create the team. This is what happened: ${error}`)
    }
    if (response.error) {
        throw new Error(response.error.message ?? 'We got an unknown error from Stripe.')
    }
    const tokenId = response.token?.id
    if (!tokenId) {
        throw new Error('Stripe token not found.')
    }
    return tokenId
}

export const CodyProCheckoutForm: React.FunctionComponent<CodyProCheckoutFormProps> = ({
    subscription,
    customerEmail,
}) => {
    const stripe = useStripe()
    const elements = useElements()
    const navigate = useNavigate()

    const [urlSearchParams] = useSearchParams()
    const addSeats = !!urlSearchParams.get('addSeats')
    const initialCurrentSeats = addSeats && subscription ? subscription.maxSeats : 0
    const maxNewSeatCount = MAX_SEAT_COUNT - initialCurrentSeats
    const initialNewSeats = Math.max(
        Math.min(maxNewSeatCount, parseInt(urlSearchParams.get('seats') || '', 10) || 1),
        MIN_SEAT_COUNT
    )
    const isTeam = addSeats || initialNewSeats > 1

    const [errorMessage, setErrorMessage] = React.useState<string | null>(null)
    // In the case of new subscriptions we have 0 initial seats, so "addedSeatCount" is actually just "seatCount".
    const [addedSeatCount, setAddedSeatCount] = React.useState(initialNewSeats)
    const [submitting, setSubmitting] = React.useState(false)

    const createTeamMutation = useCreateTeam()

    const handleSubmit = async (event: React.FormEvent<HTMLFormElement>): Promise<void> => {
        event.preventDefault()

        if (!stripe || !elements) {
            setErrorMessage('Stripe or Stripe Elements libraries not available.')
            return
        }
        const cardNumberElement = elements.getElement(CardNumberElement)
        if (!cardNumberElement) {
            setErrorMessage('CardNumberElement not found.')
            return
        }
        const addressElement = elements.getElement(AddressElement)
        if (!addressElement) {
            setErrorMessage('AddressElement not found.')
            return
        }
        const addressElementValue = await addressElement.getValue()
        if (!addressElementValue.complete) {
            setErrorMessage('Please fill out your billing address.')
            return
        }

        const suppliedAddress = addressElementValue.value.address

        setSubmitting(true)

        let token
        try {
            token = await createStripeToken(stripe, cardNumberElement, suppliedAddress)
        } catch (error) {
            setErrorMessage(error)
            setSubmitting(false)
            return
        }

        // This is where we send the token to the backend to create a subscription.
        try {
            // Even though .mutate is recommended (https://tkdodo.eu/blog/mastering-mutations-in-react-query#mutate-or-mutateasync),
            // this use makes it very convenient to just have a linear flow with error handling and a redirect at the end.
            await createTeamMutation.mutateAsync({
                name: '(no name yet)',
                slug: '(no slug yet)',
                seats: addedSeatCount,
                address: {
                    line1: suppliedAddress.line1,
                    line2: suppliedAddress.line2 || '',
                    city: suppliedAddress.city,
                    state: suppliedAddress.state,
                    postalCode: suppliedAddress.postal_code,
                    country: suppliedAddress.country,
                },
                billingInterval: 'monthly',
                couponCode: '',
                creditCardToken: token,
            })

            navigate('/cody/manage?welcome=1')

            setSubmitting(false)
        } catch (error) {
            setErrorMessage(`We couldn't create the Stripe token. This is what happened: ${error}`)
            setSubmitting(false)
        }
    }

    // TODO: Use preview prices from the backend when adding seats.
    const proRatedPrice = addedSeatCount * 9
    const priceDiff = addedSeatCount * 9
    const totalMonthlyPrice = (initialCurrentSeats + addedSeatCount) * 9

    return (
        <>
            {addedSeatCount >= 30 && (
                <CodyAlert variant="purple">
                    <H3>Explore an enterprise plan</H3>
                    <Text className="mb-0">
                        Team plans are limited to 50 users.{' '}
                        <Link to="https://sourcegraph.com/contact/sales/">Contact sales</Link> to learn more.
                    </Text>
                </CodyAlert>
            )}
            <Container>
                <Grid columnCount={2} spacing={4} className="mb-0">
                    <div>
                        <H2 className="font-medium mb-3c">{isTeam ? 'Add seats' : 'Select number of seats'}</H2>
                        <div className="d-flex flex-row align-items-center pb-3c mb-3c border-bottom">
                            <div className="flex-1">$9 per seat / month</div>
                            <Button
                                disabled={addedSeatCount === MIN_SEAT_COUNT}
                                onClick={() => setAddedSeatCount(c => (c > MIN_SEAT_COUNT ? c - 1 : c))}
                                className="px-3c py-2 border-0"
                            >
                                <Icon aria-hidden={true} svgPath={mdiMinusThick} className={styles.plusMinusButton} />
                            </Button>
                            <div className={styles.seatCountSelectorValue}>{addedSeatCount}</div>
                            <Button
                                disabled={addedSeatCount === maxNewSeatCount}
                                onClick={() => setAddedSeatCount(c => (c < maxNewSeatCount ? c + 1 : c))}
                                className="px-3c py-2 border-0"
                            >
                                <Icon aria-hidden={true} svgPath={mdiPlusThick} className={styles.plusMinusButton} />
                            </Button>
                        </div>

                        <H2 className="font-medium mb-3c">Summary</H2>
                        {addSeats && (
                            <div className="d-flex flex-row align-items-center mb-4">
                                <div className="flex-1">Pro-rated cost for this month</div>
                                <div>
                                    <strong>${proRatedPrice} / month</strong>
                                </div>
                            </div>
                        )}
                        <div className="d-flex flex-row align-items-center mb-4">
                            <div className="flex-1">
                                {addSeats ? 'Adding ' : ''} {addedSeatCount} {pluralize('seat', addedSeatCount)}
                            </div>
                            <div className={styles.price}>${priceDiff} / month</div>
                        </div>
                        {addSeats && (
                            <div className="d-flex flex-row align-items-center mb-4">
                                <div className="flex-1">
                                    Total for {initialCurrentSeats + addedSeatCount} {pluralize('seat', addedSeatCount)}
                                </div>
                                <div>
                                    <strong>${totalMonthlyPrice} / month</strong>
                                </div>
                            </div>
                        )}
                        {addSeats && (
                            <Text size="small" className={styles.disclaimer}>
                                    Each seat is pro-rated this month, and will be charged at the full rate next month.
                            </Text>
                        )}
                    </div>
                    <div>
                        <H2 className="font-medium">
                            Purchase {addedSeatCount} {pluralize('seat', addedSeatCount)}
                        </H2>
                        <Form onSubmit={handleSubmit}>
                            <StripeCardDetails className="mb-3" onFocus={() => setErrorMessage('')} />

                            <Text className="mb-2 font-medium text-sm">Email</Text>
                            <Text className="ml-3 mb-4 font-medium text-sm">{customerEmail || ''} </Text>

                            {addSeats && subscription /* TypeScript needs this */ ? (
                                <NonEditableBillingAddress subscription={subscription} />
                            ) : (
                                <StripeAddressElement onFocus={() => setErrorMessage('')} />
                            )}
                            {errorMessage && (
                                <div className={classNames(styles.paymentDataErrorMessage)}>{errorMessage}</div>
                            )}

                            <Button
                                variant="primary"
                                disabled={submitting}
                                className={classNames('d-block w-100 mb-4', styles.payButton)}
                                type="submit"
                            >
                                {submitting ? (
                                    <LoadingSpinner className={styles.lineHeightLoadingSpinner} />
                                ) : (
                                    'Subscribe'
                                )}
                            </Button>

                            <div>
                                <Text size="small" className={styles.disclaimer}>
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
        </>
    )
}
