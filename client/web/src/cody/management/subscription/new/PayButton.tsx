import React, { useCallback, type MouseEventHandler } from 'react'

import { useElements } from '@stripe/react-stripe-js'

import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import styles from './NewCodyProSubscriptionPage.module.scss'

interface PayButtonProps {
    setErrorMessage: (message: string) => void
    className?: string
    children: React.ReactNode
}

export const PayButton: React.FunctionComponent<PayButtonProps> = ({
    setErrorMessage,
    className,
    children,
    ...props
}) => {
    const elements = useElements()
    const [loading, setLoading] = React.useState(false)

    const paymentElement = elements?.getElement('payment')

    paymentElement?.('change', event => {
        if (event.error) {
            setErrorMessage(event.error.message)
        }
    })

    const handleClick: MouseEventHandler<HTMLButtonElement> = useCallback(async () => {
        if (!canConfirm) {
            if (confirmationRequirements.includes('paymentDetails')) {
                setErrorMessage('Please fill out your payment details')
            } else if (confirmationRequirements.includes('billingAddress')) {
                setErrorMessage('Please fill out your billing address')
            } else {
                setErrorMessage('Please fill out all required fields')
            }
            return
        }
        setLoading(true)
        try {
            const result = await confirm()
            if (result.error?.message) {
                setErrorMessage(result.error.message)
            }
            setLoading(false)
        } catch (error) {
            setErrorMessage(error)
            setLoading(false)
        }
    }, [canConfirm, confirm, confirmationRequirements, setErrorMessage])

    return (
        // eslint-disable-next-line no-restricted-syntax
        <Button disabled={loading} onClick={handleClick} className={className} {...props}>
            {loading ? <LoadingSpinner className={styles.lineHeightLoadingSpinner} /> : children}
        </Button>
    )
}
