import React, { useEffect, useState } from 'react'

import { mdiPencilOutline, mdiCreditCardOutline, mdiPlus, mdiCheck } from '@mdi/js'
import { CardNumberElement, Elements, useElements, useStripe } from '@stripe/react-stripe-js'
import { loadStripe } from '@stripe/stripe-js'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { Button, Form, Grid, H3, Icon, Text } from '@sourcegraph/wildcard'

import { useUpdateCurrentSubscription } from '../../api/react-query/subscriptions'
import type { PaymentMethod, Subscription } from '../../api/teamSubscriptions'
import { StripeCardDetails } from '../StripeCardDetails'

import { BillingAddress, useBillingAddressStripeElementsOptions } from './BillingAddress'
import { LoadingIconButton } from './LoadingIconButton'

import styles from './PaymentDetails.module.scss'

// NOTE: Call loadStripe outside a component’s render to avoid recreating the object.
// We do it here, meaning that "stripe.js" will get loaded lazily, when the user
// routes to this page.
const publishableKey = window.context.frontendCodyProConfig?.stripePublishableKey
if (!publishableKey) {
    logger.error('Stripe publishable key not found')
}
const stripe = await loadStripe(publishableKey || '')

const updateSubscriptionMutationErrorText =
    "We couldn't update your credit card info. Please try again. If this happens again, contact support at support@sourcegraph.com."

export const PaymentDetails: React.FC<{ subscription: Subscription }> = ({ subscription }) => {
    const options = useBillingAddressStripeElementsOptions()

    return (
        <Grid columnCount={2} spacing={0}>
            <div className={styles.gridItem}>
                <PaymentMethod paymentMethod={subscription.paymentMethod} />
            </div>
            <div className={styles.gridItem}>
                <Elements stripe={stripe} options={options}>
                    <BillingAddress subscription={subscription} title="Billing address" editable={true} />
                </Elements>
            </div>
        </Grid>
    )
}

const PaymentMethod: React.FC<{ paymentMethod: PaymentMethod | undefined }> = ({ paymentMethod }) => {
    const [isEditMode, setIsEditMode] = useState(false)

    if (!paymentMethod) {
        return <PaymentMethodMissing onAddButtonClick={() => setIsEditMode(true)} />
    }

    if (isEditMode) {
        return (
            <Elements stripe={stripe}>
                <PaymentMethodForm onReset={() => setIsEditMode(false)} onSubmit={() => setIsEditMode(false)} />
            </Elements>
        )
    }

    return <ActivePaymentMethod paymentMethod={paymentMethod} onEditButtonClick={() => setIsEditMode(true)} />
}

const PaymentMethodMissing: React.FC<{ onAddButtonClick: () => void }> = ({ onAddButtonClick }) => (
    <div className="d-flex align-items-center justify-content-between">
        <H3>No payment method is available</H3>
        <Button variant="link" className={styles.titleButton} onClick={onAddButtonClick}>
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

interface PaymentMethodFormProps {
    onReset: () => void
    onSubmit: () => void
}

const PaymentMethodForm: React.FC<PaymentMethodFormProps> = props => {
    const stripe = useStripe()
    const elements = useElements()

    const updateCurrentSubscriptionMutation = useUpdateCurrentSubscription()

    const [isStripeLoading, setIsStripeLoading] = useState(false)
    const [stripeErrorMessage, setStripeErrorMessage] = useState('')

    const [isErrorVisible, setIsErrorVisible] = useState(true)

    const isLoading = isStripeLoading || updateCurrentSubscriptionMutation.isPending
    const errorMessage =
        stripeErrorMessage || updateCurrentSubscriptionMutation.isError ? updateSubscriptionMutationErrorText : ''

    useEffect(() => {
        if (errorMessage) {
            setIsErrorVisible(true)
        }
    }, [errorMessage])

    const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (event): Promise<void> => {
        event.preventDefault()

        setStripeErrorMessage('')

        if (!stripe || !elements) {
            return setStripeErrorMessage('Stripe or Stripe Elements libraries are not available.')
        }

        const cardNumberElement = elements.getElement(CardNumberElement)
        if (!cardNumberElement) {
            return setStripeErrorMessage('Stripe card number element was not found.')
        }

        setIsStripeLoading(true)
        const tokenResult = await stripe.createToken(cardNumberElement)
        setIsStripeLoading(false)
        if (tokenResult.error) {
            return setStripeErrorMessage(tokenResult.error.message ?? 'An unknown error occurred.')
        }

        updateCurrentSubscriptionMutation.mutate(
            { customerUpdate: { newCreditCardToken: tokenResult.token.id } },
            { onSuccess: props.onSubmit }
        )
    }

    return (
        <>
            <H3>Edit credit card</H3>

            <Form onSubmit={handleSubmit} onReset={props.onReset} className={styles.paymentMethodForm}>
                <StripeCardDetails onFocus={() => setIsErrorVisible(false)} />

                {isErrorVisible && errorMessage ? <Text className="text-danger">{errorMessage}</Text> : null}

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
