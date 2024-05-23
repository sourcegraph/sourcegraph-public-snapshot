import { useMemo, useState } from 'react'

import { mdiPencilOutline, mdiCreditCardOutline, mdiPlus } from '@mdi/js'
import {
    AddressElement,
    CardCvcElement,
    CardExpiryElement,
    CardNumberElement,
    Elements,
    useElements,
    useStripe,
} from '@stripe/react-stripe-js'
import {
    loadStripe,
    type StripeCardElementOptions,
    type StripeAddressElementOptions,
    type StripeElementsOptions,
} from '@stripe/stripe-js'
import classNames from 'classnames'

import { Theme, useTheme } from '@sourcegraph/shared/src/theme'
import { Button, Form, Grid, H3, Icon, Label, Text } from '@sourcegraph/wildcard'

import type { Subscription } from '../../api/teamSubscriptions'

import styles from './PaymentDetails.module.scss'

const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey
if (!publishableKey) {
    // TODO: handle error
    throw new Error('Stripe publishable key not found')
}

const stripePromise = loadStripe(publishableKey)

export const PaymentDetails: React.FC<{ subscription: Subscription }> = ({ subscription }) => (
    <Grid columnCount={2} spacing={0} className={styles.grid}>
        <div className={styles.gridItem}>
            <PaymentMethod subscription={subscription} />
        </div>
        <div className={styles.gridItem}>
            <BillingAddress subscription={subscription} />
        </div>
    </Grid>
)

const PaymentMethod: React.FC<{ subscription: Subscription }> = ({ subscription: { paymentMethod } }) => {
    const [isEditMode, setIsEditMode] = useState(false)

    if (!paymentMethod) {
        return <PaymentMethodMissing onAddButtonClick={() => setIsEditMode(true)} />
    }

    if (isEditMode) {
        return (
            <Elements stripe={stripePromise}>
                <PaymentMethodForm onReset={() => setIsEditMode(false)} onSubmit={() => setIsEditMode(false)} />
            </Elements>
        )
    }

    return <ActivePaymentMethod paymentMethod={paymentMethod} onEditButtonClick={() => setIsEditMode(true)} />
}

const PaymentMethodMissing: React.FC<{ onAddButtonClick: () => void }> = props => (
    <div className={styles.title}>
        <H3>No payment method is available</H3>
        <Button variant="link" className={styles.titleButton} onClick={props.onAddButtonClick}>
            <Icon aria-hidden={true} svgPath={mdiPlus} className="mr-1" /> Add
        </Button>
    </div>
)

const ActivePaymentMethod: React.FC<
    Required<Pick<Subscription, 'paymentMethod'>> & { onEditButtonClick: () => void }
> = props => (
    <>
        <div className={styles.title}>
            <H3>Active credit card</H3>
            <Button variant="link" className={styles.titleButton} onClick={props.onEditButtonClick}>
                <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
            </Button>
        </div>
        <div className={styles.paymentMethodContent}>
            <Text as="span" className={classNames('text-muted', styles.paymentMethodNumber)}>
                <Icon aria-hidden={true} svgPath={mdiCreditCardOutline} /> ···· ···· ···· {props.paymentMethod.last4}
            </Text>
            <Text as="span" className="text-muted">
                Expires {props.paymentMethod.expMonth}/{props.paymentMethod.expYear}
            </Text>
        </div>
    </>
)

const useStripeCardElementOptions = (): StripeCardElementOptions => {
    const { theme } = useTheme()

    return useMemo(
        () => ({
            disableLink: true,
            hidePostalCode: true,

            classes: {
                base: classNames('form-control', styles.paymentMethodFormInput),
                focus: 'focus-visible',
                invalid: 'is-invalid',
            },

            style: {
                base: {
                    color: theme === Theme.Light ? '#262b38' : '#dbe2f0',
                },
            },
        }),
        [theme]
    )
}

const PaymentMethodForm: React.FC<{ onReset: () => void; onSubmit: () => void }> = props => {
    const stripe = useStripe()
    const elements = useElements()
    const cardElementOptions = useStripeCardElementOptions()

    const [isLoading, setIsLoading] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')

    const handleSubmit = async (): Promise<void> => {
        if (!stripe || !elements) {
            return setErrorMessage('Stripe or Stripe Elements libraries not available.')
        }

        const cardNumberElement = elements.getElement(CardNumberElement)
        if (!cardNumberElement) {
            return setErrorMessage('CardNumber element was not found.')
        }

        const tokenResult = await stripe.createToken(cardNumberElement)
        if (tokenResult.error) {
            return setErrorMessage(tokenResult.error.message ?? 'An unknown error occurred.')
        }

        setIsLoading(true)
        try {
            // TODO: call SSC API
            props.onSubmit()
        } catch (error) {
            // TODO[accounts.sourcegraph.com#353]: Send error to Sentry
            // eslint-disable-next-line no-console
            console.error(error)

            // // If there is a human-friendly error in the GraphQL response, surface that to the user.
            // const apolloError = error as ApolloError
            // if (apolloError.name === 'ApolloError') {
            //     if (apolloError.message !== 'Internal Server Error') {
            //         setErrorMessage(
            //             `An error occurred while updating your credit card information: ${apolloError.message}`
            //         )
            //         return
            //     }
            // }
            setErrorMessage(
                'An error occurred while updating your credit card info. Please try again. If the problem persists, contact support at support@sourcegraph.com.'
            )
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <>
            <H3>Edit credit card</H3>

            <Form onSubmit={handleSubmit} onReset={props.onReset} className={styles.paymentMethodForm}>
                <div>
                    <Label className={styles.paymentMethodFormLabel}>
                        <Text className="mb-2">Card number</Text>
                        <CardNumberElement options={cardElementOptions} onFocus={() => {}} />
                    </Label>
                </div>

                <Grid columnCount={2} className="mt-3 mb-0 pb-3">
                    <Label className={styles.paymentMethodFormLabel}>
                        <Text className="mb-2">Expiry date</Text>
                        <CardExpiryElement options={cardElementOptions} onFocus={() => {}} />
                    </Label>

                    <Label className={styles.paymentMethodFormLabel}>
                        <Text className="mb-2">CVC</Text>
                        <CardCvcElement options={cardElementOptions} onFocus={() => {}} />
                    </Label>
                </Grid>

                {errorMessage && <Text className="text-danger">{errorMessage}</Text>}

                <div className={classNames('mt-4', styles.paymentMethodFormButtonContainer)}>
                    <Button type="reset" variant="secondary" outline={true}>
                        Cancel
                    </Button>
                    <Button disabled={isLoading} type="submit" variant="primary" className="ml-2">
                        Save
                    </Button>
                </div>
            </Form>
        </>
    )
}

const useBillingAddressStripeElementsOptions = (): StripeElementsOptions => {
    const { theme } = useTheme()

    return useMemo(
        () => ({
            appearance: {
                variables: {
                    // corresponds to var(--font-family-base)
                    fontFamily:
                        "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji'",
                    gridRowSpacing: '16px',
                    borderRadius: '3px',
                },

                rules: {
                    '.Label': {
                        marginBottom: '8px',
                        fontWeight: '500',
                        color: theme === Theme.Light ? '#343a4d' : '#dbe2f0',
                        lineHeight: '20px',
                        fontSize: '14px',
                    },
                    '.Input': {
                        backgroundColor: theme === Theme.Light ? '#ffffff' : '#1d212f',
                        color: theme === Theme.Light ? '#262b38' : '#dbe2f0',
                        paddingTop: '6px',
                        paddingBottom: '6px',
                        borderColor: theme === Theme.Light ? '#dbe2f0' : '#343a4d',
                        boxShadow: 'none',
                        lineHeight: '20px',
                        fontSize: '14px',
                    },
                    '.Input:focus': {
                        borderColor: '#0b70db',
                        boxShadow: `0 0 0 0.125rem ${theme === Theme.Light ? '#a3d0ff' : '#0f59aa'}`,
                    },
                },
            },
        }),
        [theme]
    )
}

const BillingAddress: React.FC<{ subscription: Subscription }> = ({ subscription }) => {
    const options = useBillingAddressStripeElementsOptions()
    const [isEditMode, setIsEditMode] = useState(false)

    return (
        <div>
            <div className={styles.title}>
                <H3>Billing address</H3>
                <Button variant="link" className={styles.titleButton} onClick={() => setIsEditMode(true)}>
                    <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
                </Button>
            </div>

            {isEditMode ? (
                <Elements stripe={stripePromise} options={options}>
                    <BillingAddressForm
                        subscription={subscription}
                        onReset={() => setIsEditMode(false)}
                        onSubmit={() => setIsEditMode(false)}
                    />
                </Elements>
            ) : (
                <ActiveBillingAddress subscription={subscription} />
            )}
        </div>
    )
}

const ActiveBillingAddress: React.FC<{ subscription: Subscription }> = ({ subscription }) => (
    <div>
        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Full name
            </Text>
            <Text className="font-weight-medium">{subscription.name}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Country or region
            </Text>
            <Text className="font-weight-medium">{subscription.address.country || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Address line 1
            </Text>
            <Text className="font-weight-medium">{subscription.address.line1 || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Address line 2
            </Text>
            <Text className="font-weight-medium">{subscription.address.line2 || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                City
            </Text>
            <Text className="font-weight-medium">{subscription.address.city || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                State
            </Text>
            <Text className="font-weight-medium">{subscription.address.state || '-'}</Text>
        </div>

        <div className="mt-3">
            <Text size="small" className="mb-1 text-muted font-weight-medium">
                Postal code
            </Text>
            <Text className="font-weight-medium">{subscription.address.postalCode || '-'}</Text>
        </div>
    </div>
)

const BillingAddressForm: React.FC<{
    subscription: Subscription
    onReset: () => void
    onSubmit: () => void
}> = ({ subscription, onReset, onSubmit }) => {
    const stripe = useStripe()
    const elements = useElements()

    const [isLoading, setIsLoading] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')

    const handleSubmit = async (event): Promise<void> => {
        onSubmit()
        return
    }

    const options: StripeAddressElementOptions = useMemo(
        () => ({
            mode: 'billing',
            display: { name: 'full' },
            defaultValues: {
                name: subscription.name,
                address: {
                    line1: subscription.address.line1,
                    line2: subscription.address.line2,
                    city: subscription.address.city,
                    state: subscription.address.state,
                    postal_code: subscription.address.postalCode,
                    country: subscription.address.country,
                },
            },
        }),
        [subscription]
    )

    return (
        <Form onSubmit={handleSubmit} onReset={onReset} className={styles.billingAddressForm}>
            <AddressElement
                options={options}
                onFocus={() => {
                    // setErrorMessage(null)
                }}
            />

            <div className={styles.billingAddressFormButtonContainer}>
                <Button type="reset" variant="secondary" outline={true}>
                    Cancel
                </Button>
                <Button disabled={isLoading} type="submit" variant="primary" className="ml-2">
                    Save
                </Button>
            </div>
        </Form>
    )
}
