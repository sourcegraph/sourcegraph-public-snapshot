import React, { useEffect, useState } from 'react'

import { mdiCheck } from '@mdi/js'
import { CardNumberElement, Elements, useElements, useStripe } from '@stripe/react-stripe-js'
import { loadStripe } from '@stripe/stripe-js'

import { logger } from '@sourcegraph/common'
import { Button, Form, Grid, H3, Text } from '@sourcegraph/wildcard'

import { useUpdateCurrentSubscription } from '../../api/react-query/subscriptions'
// Suppressing false positive caused by an ESLint bug. See https://github.com/typescript-eslint/typescript-eslint/issues/4608
// eslint-disable-next-line @typescript-eslint/consistent-type-imports
import type { PaymentMethod, Subscription } from '../../api/types'
import { PaymentMethodPreview } from '../PaymentMethodPreview'
import { StripeCardDetails } from '../StripeCardDetails'

import { BillingAddress } from './BillingAddress'
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

export const PaymentDetails: React.FC<{ subscription: Subscription }> = ({ subscription }) => (
    <Grid columnCount={2} spacing={0}>
        <div className={styles.gridItem}>
            <PaymentMethod paymentMethod={subscription.paymentMethod} />
        </div>
        <div className={styles.gridItem}>
            <BillingAddress stripe={stripe} subscription={subscription} />
        </div>
    </Grid>
)

const PaymentMethod: React.FC<{ paymentMethod: PaymentMethod | undefined }> = ({ paymentMethod }) => {
    const [isEditMode, setIsEditMode] = useState(false)

    if (isEditMode && paymentMethod) {
        return (
            <Elements stripe={stripe}>
                <PaymentMethodForm onReset={() => setIsEditMode(false)} onSubmit={() => setIsEditMode(false)} />
            </Elements>
        )
    }

    return (
        <PaymentMethodPreview
            paymentMethod={paymentMethod}
            isEditable={true}
            onButtonClick={() => setIsEditMode(true)}
        />
    )
}

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
