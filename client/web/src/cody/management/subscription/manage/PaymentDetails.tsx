import { useContext, useMemo, useState } from 'react'

import { mdiPencilOutline, mdiCreditCardOutline, mdiPlus, mdiCheck } from '@mdi/js'
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

import { logger } from '@sourcegraph/common'
import { Theme, useTheme } from '@sourcegraph/shared/src/theme'
import { Button, Form, Grid, H3, Icon, Label, Text } from '@sourcegraph/wildcard'

import { Client } from '../../api/client'
import { CodyProApiClientContext } from '../../api/components/CodyProApiClient'
import type { Subscription } from '../../api/teamSubscriptions'

import { LoadingIconButton } from './LoadingIconButton'

import styles from './PaymentDetails.module.scss'

const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey
if (!publishableKey) {
    logger.error('Stripe publishable key not found')
}

const stripePromise = loadStripe(publishableKey || '')

interface PaymentDetailsProps {
    subscription: Subscription
    refetchSubscription: () => Promise<void>
}

export const PaymentDetails: React.FC<PaymentDetailsProps> = props => (
    <Grid columnCount={2} spacing={0}>
        <div className={styles.gridItem}>
            <PaymentMethod subscription={props.subscription} refetchSubscription={props.refetchSubscription} />
        </div>
        <div className={styles.gridItem}>
            <BillingAddress subscription={props.subscription} refetchSubscription={props.refetchSubscription} />
        </div>
    </Grid>
)

const PaymentMethod: React.FC<PaymentDetailsProps> = props => {
    const [isEditMode, setIsEditMode] = useState(false)

    if (!props.subscription.paymentMethod) {
        return <PaymentMethodMissing onAddButtonClick={() => setIsEditMode(true)} />
    }

    if (isEditMode) {
        return (
            <Elements stripe={stripePromise}>
                <PaymentMethodForm
                    onReset={() => setIsEditMode(false)}
                    onSubmit={() => setIsEditMode(false)}
                    refetchSubscription={props.refetchSubscription}
                />
            </Elements>
        )
    }

    return (
        <ActivePaymentMethod
            paymentMethod={props.subscription.paymentMethod}
            onEditButtonClick={() => setIsEditMode(true)}
        />
    )
}

const PaymentMethodMissing: React.FC<{ onAddButtonClick: () => void }> = props => (
    <div className="d-flex align-items-center justify-content-between">
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
        <div className="d-flex align-items-center justify-content-between">
            <H3>Active credit card</H3>
            <Button variant="link" className={styles.titleButton} onClick={props.onEditButtonClick}>
                <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
            </Button>
        </div>
        <div className="mt-3 d-flex justify-content-between">
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
                base: 'form-control py-2',
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

interface PaymentMethodFormProps extends Pick<PaymentDetailsProps, 'refetchSubscription'> {
    onReset: () => void
    onSubmit: () => void
}

const PaymentMethodForm: React.FC<PaymentMethodFormProps> = props => {
    const stripe = useStripe()
    const elements = useElements()
    const cardElementOptions = useStripeCardElementOptions()

    const { caller } = useContext(CodyProApiClientContext)

    const [isLoading, setIsLoading] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')

    const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (event): Promise<void> => {
        event.preventDefault()

        setIsLoading(true)
        setErrorMessage('')

        if (!stripe || !elements) {
            setIsLoading(false)
            return setErrorMessage('Stripe or Stripe Elements libraries are not available.')
        }

        const cardNumberElement = elements.getElement(CardNumberElement)
        if (!cardNumberElement) {
            setIsLoading(false)
            return setErrorMessage('Stripe card number element was not found.')
        }

        const tokenResult = await stripe.createToken(cardNumberElement)
        if (tokenResult.error) {
            setIsLoading(false)
            return setErrorMessage(tokenResult.error.message ?? 'An unknown error occurred.')
        }

        const serverErrorText =
            'An error occurred while updating your credit card info. Please try again. If the problem persists, contact support at support@sourcegraph.com.'

        try {
            const { response } = await caller.call(
                Client.updateCurrentSubscription({ customerUpdate: { newCreditCardToken: tokenResult.token.id } })
            )
            if (response.ok) {
                await props.refetchSubscription()
                props.onSubmit()
            } else {
                setErrorMessage(serverErrorText)
            }
        } catch (error) {
            logger.error(error)
            setErrorMessage(serverErrorText)
        } finally {
            setIsLoading(false)
        }
    }

    const cardElementProps = { options: cardElementOptions, onFocus: () => setErrorMessage('') }

    return (
        <>
            <H3>Edit credit card</H3>

            <Form onSubmit={handleSubmit} onReset={props.onReset} className={styles.paymentMethodForm}>
                <div>
                    <Label className="d-block">
                        <Text className="mb-2">Card number</Text>
                        <CardNumberElement {...cardElementProps} />
                    </Label>
                </div>

                <Grid columnCount={2} className="mt-3 mb-0 pb-3">
                    <Label className="d-block">
                        <Text className="mb-2">Expiry date</Text>
                        <CardExpiryElement {...cardElementProps} />
                    </Label>

                    <Label className="d-block">
                        <Text className="mb-2">CVC</Text>
                        <CardCvcElement {...cardElementProps} />
                    </Label>
                </Grid>

                {errorMessage && <Text className="text-danger">{errorMessage}</Text>}

                <div className="mt-4 d-flex justify-content-end">
                    <Button type="reset" variant="secondary" outline={true}>
                        Cancel
                    </Button>
                    <LoadingIconButton
                        type="submit"
                        variant="primary"
                        className="ml-2"
                        disabled={isLoading}
                        isLoading={isLoading}
                        iconSvgPath={mdiCheck}
                    >
                        Save
                    </LoadingIconButton>
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

const BillingAddress: React.FC<PaymentDetailsProps> = props => {
    const options = useBillingAddressStripeElementsOptions()
    const [isEditMode, setIsEditMode] = useState(false)

    return (
        <div>
            <div className="d-flex align-items-center justify-content-between">
                <H3>Billing address</H3>
                <Button variant="link" className={styles.titleButton} onClick={() => setIsEditMode(true)}>
                    <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
                </Button>
            </div>

            {isEditMode ? (
                <Elements stripe={stripePromise} options={options}>
                    <BillingAddressForm
                        subscription={props.subscription}
                        refetchSubscription={props.refetchSubscription}
                        onReset={() => setIsEditMode(false)}
                        onSubmit={() => setIsEditMode(false)}
                    />
                </Elements>
            ) : (
                <ActiveBillingAddress subscription={props.subscription} />
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

interface BillingAddressFormProps extends PaymentDetailsProps {
    onReset: () => void
    onSubmit: () => void
}

const BillingAddressForm: React.FC<BillingAddressFormProps> = props => {
    const stripe = useStripe()
    const elements = useElements()

    const { caller } = useContext(CodyProApiClientContext)

    const [isLoading, setIsLoading] = useState(false)
    const [errorMessage, setErrorMessage] = useState('')

    const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (event): Promise<void> => {
        event.preventDefault()

        setIsLoading(true)
        setErrorMessage('')

        if (!stripe || !elements) {
            setIsLoading(false)
            return setErrorMessage('Stripe or Stripe Elements libraries are not available.')
        }

        const addressElement = elements.getElement(AddressElement)
        if (!addressElement) {
            setIsLoading(false)
            return setErrorMessage('Stripe address element was not found.')
        }

        const addressElementValue = await addressElement.getValue()
        if (!addressElementValue.complete) {
            setIsLoading(false)
            return setErrorMessage('Address is not complete.')
        }

        const serverErrorText =
            'An error occurred while updating your credit card info. Please try again. If the problem persists, contact support at support@sourcegraph.com.'

        try {
            const { line1, line2, postal_code, city, state, country } = addressElementValue.value.address
            const { response } = await caller.call(
                Client.updateCurrentSubscription({
                    customerUpdate: {
                        newName: addressElementValue.value.name,
                        newAddress: {
                            line1,
                            line2: line2 || '',
                            postalCode: postal_code,
                            city,
                            state,
                            country,
                        },
                    },
                })
            )
            if (response.ok) {
                await props.refetchSubscription()
                props.onSubmit()
            } else {
                setErrorMessage(serverErrorText)
            }
        } catch (error) {
            logger.error(error)
            setErrorMessage(serverErrorText)
        } finally {
            setIsLoading(false)
        }
    }

    const options = useMemo(
        (): StripeAddressElementOptions => ({
            mode: 'billing',
            display: { name: 'full' },
            defaultValues: {
                name: props.subscription.name,
                address: {
                    ...props.subscription.address,
                    postal_code: props.subscription.address.postalCode,
                },
            },
        }),
        [props.subscription]
    )

    return (
        <Form onSubmit={handleSubmit} onReset={props.onReset} className={styles.billingAddressForm}>
            <AddressElement options={options} onFocus={() => setErrorMessage('')} />

            {errorMessage && <Text className="mt-3 text-danger">{errorMessage}</Text>}

            <div className={classNames('d-flex justify-content-end', styles.billingAddressFormButtonContainer)}>
                <Button type="reset" variant="secondary" outline={true}>
                    Cancel
                </Button>
                <LoadingIconButton
                    type="submit"
                    variant="primary"
                    className="ml-2"
                    disabled={isLoading}
                    isLoading={isLoading}
                    iconSvgPath={mdiCheck}
                >
                    Save
                </LoadingIconButton>
            </div>
        </Form>
    )
}
