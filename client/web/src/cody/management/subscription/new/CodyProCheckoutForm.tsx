import React from 'react'

import { mdiMinusThick, mdiPlusThick } from '@mdi/js'
import { AddressElement, useStripe, useElements, CardNumberElement } from '@stripe/react-stripe-js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { pluralize } from '@sourcegraph/common'
import {
    Form,
    Link,
    Button,
    Grid,
    H2,
    Text,
    Container,
    Icon,
    Input,
    Label,
    H3,
    LoadingSpinner,
} from '@sourcegraph/wildcard'

import { CodyAlert } from '../../../components/CodyAlert'
import { useCreateTeam } from '../../api/react-query/subscriptions'
import { StripeAddressElement } from '../StripeAddressElement'
import { StripeCardDetails } from '../StripeCardDetails'

import styles from './NewCodyProSubscriptionPage.module.scss'

const MIN_SEAT_COUNT = 1
const MAX_SEAT_COUNT = 50

interface CodyProCheckoutFormProps {
    initialSeatCount: number
    customerEmail: string | undefined
}

export const CodyProCheckoutForm: React.FunctionComponent<CodyProCheckoutFormProps> = ({
    initialSeatCount,
    customerEmail,
}) => {
    const stripe = useStripe()
    const elements = useElements()
    const navigate = useNavigate()

    const isTeam = initialSeatCount > 1

    const [errorMessage, setErrorMessage] = React.useState<string | null>(null)
    const [seatCount, setSeatCount] = React.useState(initialSeatCount)
    const [submitting, setSubmitting] = React.useState(false)

    // N * $9. We expect this to be more complex later with annual plans, etc.
    const total = seatCount * 9

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
        try {
            // Note that Stripe may have returned an error response.
            const response = await stripe.createToken(cardNumberElement, {
                // We want to include the address data along with the card info so Stripe can do
                // more validation such as confirming the zip code matches the card's.
                //
                // This is information we'll also want to pass along to the backend, so
                // we can store it as the Customer's address as well.
                address_line1: suppliedAddress.line1,
                address_line2: suppliedAddress.line2 || '',
                address_city: suppliedAddress.city,
                address_state: suppliedAddress.state,
                address_zip: suppliedAddress.postal_code,
                address_country: suppliedAddress.country,
                currency: 'usd',
            })
            if (response.error) {
                setErrorMessage(response.error.message ?? 'We got an unknown error from Stripe.')
                setSubmitting(false)
                return
            }
            const token = response.token?.id
            if (!token) {
                setErrorMessage('Stripe token not found.')
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
                    seats: seatCount,
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
            } catch (error) {
                setErrorMessage(`We couldn't create the team. This is what happened: ${error}`)
                setSubmitting(false)
                return
            }

            navigate('/cody/manage?welcome=1')

            setSubmitting(false)
        } catch (error) {
            setErrorMessage(`We couldn't create the Stripe token. This is what happened: ${error}`)
            setSubmitting(false)
        }
    }

    return (
        <>
            {seatCount >= 30 && (
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
                        <H2>{isTeam ? 'Add seats' : 'Select number of seats'}</H2>
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
                                {isTeam ? 'Adding ' : ''} {seatCount} {pluralize('seat', seatCount)}
                            </div>
                            <div>
                                <strong>${total} / month</strong>
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
                        <Input value={customerEmail || ''} disabled={true} className="mb-4" />
                        <Form onSubmit={handleSubmit}>
                            <StripeCardDetails className="mb-4" onFocus={() => setErrorMessage('')} />
                            <StripeAddressElement onFocus={() => setErrorMessage('')} />
                            {errorMessage && (
                                <div className={classNames(styles.paymentDataErrorMessage)}>{errorMessage}</div>
                            )}

                            <Button
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
        </>
    )
}
