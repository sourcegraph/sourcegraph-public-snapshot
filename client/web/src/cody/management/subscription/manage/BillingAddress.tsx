import React, { useMemo, useState, useEffect } from 'react'

import { mdiPencilOutline, mdiCheck } from '@mdi/js'
import { useStripe, useElements, AddressElement, Elements } from '@stripe/react-stripe-js'
import type { Stripe, StripeElementsOptions } from '@stripe/stripe-js'
import classNames from 'classnames'

import { useTheme, Theme } from '@sourcegraph/shared/src/theme'
import { H3, Button, Icon, Text, Form } from '@sourcegraph/wildcard'

import { useUpdateCurrentSubscription } from '../../api/react-query/subscriptions'
import type { Subscription } from '../../api/teamSubscriptions'
import { StripeAddressElement } from '../StripeAddressElement'

import { LoadingIconButton } from './LoadingIconButton'
import { NonEditableBillingAddress } from './NonEditableBillingAddress'

import styles from './PaymentDetails.module.scss'

export const useBillingAddressStripeElementsOptions = (): StripeElementsOptions => {
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

interface BillingAddressProps {
    stripe: Stripe | null
    subscription: Subscription
    title?: string
    editable: boolean
}

export const BillingAddress: React.FC<BillingAddressProps> = ({ stripe, subscription, title, editable }) => {
    const [isEditMode, setIsEditMode] = useState(false)

    const options = useBillingAddressStripeElementsOptions()

    return (
        <div>
            <div className="d-flex align-items-center justify-content-between">
                {title ?? <H3>{title}</H3>}
                {editable && (
                    <Button variant="link" className={styles.titleButton} onClick={() => setIsEditMode(true)}>
                        <Icon aria-hidden={true} svgPath={mdiPencilOutline} className="mr-1" /> Edit
                    </Button>
                )}
            </div>

            {isEditMode ? (
                <Elements stripe={stripe} options={options}>
                    <BillingAddressForm
                        subscription={subscription}
                        onReset={() => setIsEditMode(false)}
                        onSubmit={() => setIsEditMode(false)}
                    />
                </Elements>
            ) : (
                <NonEditableBillingAddress subscription={subscription} />
            )}
        </div>
    )
}

interface BillingAddressFormProps {
    subscription: Subscription
    onReset: () => void
    onSubmit: () => void
}

const updateSubscriptionMutationErrorText =
    "We couldn't update your credit card info. Please try again. If this happens again, contact support at support@sourcegraph.com."

const BillingAddressForm: React.FC<BillingAddressFormProps> = ({ subscription, onReset, onSubmit }) => {
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

        const addressElement = elements.getElement(AddressElement)
        if (!addressElement) {
            return setStripeErrorMessage('Stripe address element was not found.')
        }

        setIsStripeLoading(true)
        const addressElementValue = await addressElement.getValue()
        setIsStripeLoading(false)
        if (!addressElementValue.complete) {
            return setStripeErrorMessage('Address is not complete.')
        }

        const { line1, line2, postal_code, city, state, country } = addressElementValue.value.address
        updateCurrentSubscriptionMutation.mutate(
            {
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
            },
            { onSuccess: onSubmit }
        )
    }

    return (
        <Form onSubmit={handleSubmit} onReset={onReset} className={styles.billingAddressForm}>
            <StripeAddressElement subscription={subscription} onFocus={() => setIsErrorVisible(false)} />

            {isErrorVisible && errorMessage ? <Text className="mt-3 text-danger">{errorMessage}</Text> : null}

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
