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
    type Appearance,
    loadStripe,
    type StripeCardElementOptions,
    type StripeAddressElementOptions,
} from '@stripe/stripe-js'
import classNames from 'classnames'

import { Button, Form, Grid, H3, Icon, Label, Text } from '@sourcegraph/wildcard'

import type { Subscription } from '../../api/teamSubscriptions'

import styles from './PaymentDetails.module.scss'

const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey
if (!publishableKey) {
    // TODO: handle error
    throw new Error('Stripe publishable key not found')
}

const stripePromise = loadStripe(publishableKey)

const appearance: Appearance = {
    theme: 'stripe',
    variables: {
        colorPrimary: '#00b4d9',
    },
    rules: {
        '.Label': {
            marginBottom: '8px',
            fontWeight: '500',
            color: '#343a4d',
            fontSize: '14px',
            lineHeight: '20px',
        },
        '.Input': {
            marginBottom: '4px',
            paddingTop: '6px',
            paddingBottom: '6px',
            borderColor: '#dbe2f0',
            borderRadius: '3px',
            fontSize: '14px',
            lineHeight: '20px',
            boxShadow: 'none',
        },
        '.Input:focus': {
            borderColor: '#0b70db',
            boxShadow: '0 0 0 0.125rem #a3d0ff',
        },
    },
}

const noop = (): void => {}

export const PaymentDetails: React.FC<{ subscription: Subscription }> = ({ subscription }) => (
    <Elements stripe={stripePromise} options={{ appearance }}>
        <Grid columnCount={2} spacing={0} className={styles.grid}>
            <div className={styles.gridItem}>
                <PaymentMethod subscription={subscription} onChange={noop} />
            </div>
            <div className={styles.gridItem}>
                <BillingAddress subscription={subscription} onChange={noop} />
            </div>
        </Grid>
    </Elements>
)

const PaymentMethod: React.FC<{
    subscription: Subscription
    onChange: () => unknown
}> = ({ subscription: { paymentMethod }, onChange }) => {
    const [isEditMode, setIsEditMode] = useState(false)

    if (!paymentMethod) {
        return <PaymentMethodMissing onAddButtonClick={() => setIsEditMode(true)} />
    }

    if (isEditMode) {
        return <PaymentMethodForm onReset={() => setIsEditMode(false)} onSubmit={() => setIsEditMode(false)} />
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

const cardElementOptions: StripeCardElementOptions = {
    // Don't use Stripe Link. Just the basics.
    disableLink: true,
    // Since it is supplied by the AddressElement.
    hidePostalCode: true,

    // apply default wildcard input classes
    classes: {
        base: 'form-control',
        focus: 'focus-visible',
        invalid: 'is-invalid',
    },
}

const PaymentMethodForm: React.FC<{ onReset: () => void; onSubmit: () => void }> = props => {
    const stripe = useStripe()
    const elements = useElements()

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

const BillingAddress: React.FC<{
    subscription: Subscription
    onChange: () => unknown
}> = ({ subscription, onChange }) => {
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
                <BillingAddressForm
                    subscription={subscription}
                    onReset={() => setIsEditMode(false)}
                    onSubmit={() => setIsEditMode(false)}
                />
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
}> = props => {
    const stripe = useStripe()
    const elements = useElements()

    const [isLoading, setIsLoading] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')

    const handleSubmit = async (event): Promise<void> => {
        props.onSubmit()
        return
    }

    // TODO: Customize this further, enabling validation, default forms, autocomplete, etc.
    // https://stripe.com/docs/js/elements_object/create_address_element
    const options: StripeAddressElementOptions = {
        mode: 'billing',
        display: { name: 'full' },
        defaultValues: {
            name: props.subscription.name,
            address: {
                line1: props.subscription.address.line1,
                line2: props.subscription.address.line2,
                city: props.subscription.address.city,
                state: props.subscription.address.state,
                postal_code: props.subscription.address.postalCode,
                country: props.subscription.address.country,
            },
        },
    }

    return (
        <Form onSubmit={handleSubmit} onReset={props.onReset} className={styles.billingAddressForm}>
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
